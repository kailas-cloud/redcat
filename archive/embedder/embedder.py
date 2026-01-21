import os
import json
import pandas as pd
from fastembed import TextEmbedding


# -----------------------------------------------------------
# ENV CONFIG
# -----------------------------------------------------------

def load_env():
    input_path = os.getenv("FS_CATEGORIES_RAW_DATA")
    output_path = os.getenv("EMBEDDER_OUT")
    model_name = os.getenv("EMBEDDER_MODEL")

    if not input_path:
        raise RuntimeError("FS_CATEGORIES_RAW_DATA is not set")
    if not output_path:
        raise RuntimeError("EMBEDDER_OUT is not set")
    if not model_name:
        raise RuntimeError("EMBEDDER_MODEL is not set")

    return input_path, output_path, model_name


# -----------------------------------------------------------
# LOAD DATA
# -----------------------------------------------------------

def load_categories(input_path: str) -> pd.DataFrame:
    print(f"Reading parquet: {input_path}")
    df = pd.read_parquet(input_path)

    df = df[[
        "category_id",
        "category_name",
        "category_level",
        "category_label"
    ]].dropna(subset=["category_id", "category_label"])

    print(f"Loaded {len(df)} rows")
    return df


# -----------------------------------------------------------
# LOAD MODEL
# -----------------------------------------------------------

def load_model(model_name: str) -> TextEmbedding:
    print(f"Loading embedding model: {model_name}")
    return TextEmbedding(model_name=model_name)


# -----------------------------------------------------------
# EMBEDDINGS
# -----------------------------------------------------------

def compute_embeddings(embedder: TextEmbedding, labels: list) -> list:
    print("Computing embeddings…")
    inputs = [f"passage: place category: {label}" for label in labels]
    vectors = list(embedder.embed(inputs))
    print(f"Got {len(vectors)} embeddings")
    return vectors


# -----------------------------------------------------------
# BUILD JSON STRUCTURE
# -----------------------------------------------------------

def build_records(df: pd.DataFrame, embeddings: list) -> list:
    print("Building final records…")
    records = []

    for (_, row), emb in zip(df.iterrows(), embeddings):
        records.append({
            "category_id": row["category_id"],
            "category_name": row["category_name"],
            "category_level": int(row["category_level"]),
            "category_label": row["category_label"],
            "embedding": list(emb)
        })

    return records


# -----------------------------------------------------------
# SAVE JSON
# -----------------------------------------------------------

def save_json(records: list, output_path: str):
    print(f"Saving JSON: {output_path}")
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(records, f, ensure_ascii=False, indent=2)
    print("Done.")


# -----------------------------------------------------------
# MAIN
# -----------------------------------------------------------

def main():
    input_path, output_path, model_name = load_env()
    df = load_categories(input_path)
    embedder = load_model(model_name)
    embeddings = compute_embeddings(embedder, df["category_label"].tolist())
    records = build_records(df, embeddings)
    save_json(records, output_path)


if __name__ == "__main__":
    main()
