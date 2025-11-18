import os
import json
import pandas as pd
import numpy as np
from fastembed import TextEmbedding


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
    vectors = np.array([r["embedding"] for r in data])
    ids = [r["category_id"] for r in data]

    print(f"Using model: {model_name}")
    embedder = TextEmbedding(model_name=model_name)

    query = "ирландский паб"
    print(f"Query: {query}")
    q_vec = np.array(list(embedder.embed([query]))[0])

    # cosine similarity
    dot = vectors @ q_vec
    norm = np.linalg.norm(vectors, axis=1) * np.linalg.norm(q_vec)
    scores = dot / (norm + 1e-9)

    topk = np.argsort(-scores)[:10]

    for i in topk:
        print(f"{scores[i]:.4f}  {ids[i]}  {labels[i]}")


if __name__ == "__main__":
    main()
