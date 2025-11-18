import os
import json
import pandas as pd
import numpy as np
import warnings
from fastembed import TextEmbedding

warnings.filterwarnings('ignore', category=UserWarning)
warnings.filterwarnings('ignore', message='.*urllib3 v2.*')
warnings.filterwarnings('ignore', message='.*NotOpenSSLWarning.*')


QUERY = "craft beer"


def main():
    input_path = os.getenv("FS_CATEGORIES_RAW_DATA")
    output_path = os.getenv("FS_CATEGORIES_EMBEDDINGS_OUT", "category_embeddings.json")
    model_name = os.getenv("EMBEDDER_MODEL", "intfloat/multilingual-e5-large")

    if not input_path:
        raise RuntimeError("FS_CATEGORIES_RAW_DATA is not set")

    # --- источник ---
    print(f"Reading source parquet: {input_path}")
    src = pd.read_parquet(input_path)
    src = src[[
        "category_id",
        "category_name",
        "category_level",
        "category_label",
    ]].dropna(subset=["category_id", "category_label"])

    src = src.drop_duplicates(subset=["category_id"])
    print(f"Source unique categories: {len(src)}")

    # --- результат ---
    print(f"Reading embeddings json: {output_path}")
    with open(output_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    print(f"JSON records: {len(data)}")

    ids_json = {r["category_id"] for r in data}
    ids_src = set(src["category_id"])

    print(f"Unique ids in src:  {len(ids_src)}")
    print(f"Unique ids in json: {len(ids_json)}")

    missing_in_json = ids_src - ids_json
    extra_in_json = ids_json - ids_src

    print(f"Missing in json: {len(missing_in_json)}")
    print(f"Extra in json:   {len(extra_in_json)}")

    if missing_in_json or extra_in_json:
        print("❌ ID mismatch detected")
    else:
        print("✅ IDs match")

    # --- быстрый поиск ---
    print("\n=== test search ===")
    labels = [r["category_label"] for r in data]

    # Convert to float64 first to avoid overflow during conversion
    vectors = np.array([r["embedding"] for r in data], dtype=np.float64)
    ids = [r["category_id"] for r in data]

    # Check for invalid values in embeddings
    print(f"Vector shape: {vectors.shape}")
    print(f"Vector dtype: {vectors.dtype}")
    print(f"Min value: {np.min(vectors)}")
    print(f"Max value: {np.max(vectors)}")
    print(f"NaN count: {np.sum(np.isnan(vectors))}")
    print(f"Inf count: {np.sum(np.isinf(vectors))}")

    if np.any(np.isnan(vectors)) or np.any(np.isinf(vectors)):
        print("⚠️  Warning: Found NaN or inf values in embeddings - cleaning...")
        vectors = np.nan_to_num(vectors, nan=0.0, posinf=1.0, neginf=-1.0)

    print(f"Using model: {model_name}")
    embedder = TextEmbedding(model_name=model_name)

    print(f"Query: {QUERY}")

    q_vec = list(embedder.embed([f"query: {QUERY}"]))[0]
    q_vec = np.asarray(q_vec, dtype=np.float64)

    print(f"\nQuery vector stats:")
    print(f"  Shape: {q_vec.shape}")
    print(f"  Min: {np.min(q_vec):.6f}")
    print(f"  Max: {np.max(q_vec):.6f}")
    print(f"  Norm: {np.linalg.norm(q_vec):.6f}")

    # Normalize vectors safely
    v_norms = np.linalg.norm(vectors, axis=1)
    print(f"\nDocument vector norms:")
    print(f"  Min norm: {np.min(v_norms):.6f}")
    print(f"  Max norm: {np.max(v_norms):.6f}")
    print(f"  Zero norms: {np.sum(v_norms < 1e-9)}")

    # Avoid division by zero
    v_norms = np.where(v_norms < 1e-9, 1.0, v_norms)
    vectors_norm = vectors / v_norms[:, np.newaxis]

    q_norm = np.linalg.norm(q_vec)
    if q_norm < 1e-9:
        print("⚠️  Warning: Query vector has zero norm")
        q_vec_norm = q_vec
    else:
        q_vec_norm = q_vec / q_norm

    # Compute cosine similarity
    scores = np.dot(vectors_norm, q_vec_norm)

    # Clip scores to valid range [-1, 1]
    scores = np.clip(scores, -1.0, 1.0)

    topk = np.argsort(-scores)[:10]

    print("\nTop 10 matches:")
    for i in topk:
        print(f"{scores[i]:.4f}  {ids[i]}  {labels[i]}")


if __name__ == "__main__":
    main()