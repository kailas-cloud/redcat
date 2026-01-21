import json
import os
import logging

import redis


def get_redis():
    return redis.Redis(
        host=os.getenv("REDIS_HOST", "localhost"),
        port=int(os.getenv("REDIS_PORT", "6379")),
        db=int(os.getenv("REDIS_DB", "0")),
        password=os.getenv("REDIS_PASSWORD") or None,
    )


def load_categories(path: str = "categories.json"):
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def migrate():
    logging.basicConfig(level=logging.INFO, format="%(message)s")

    r = get_redis()
    categories = load_categories(os.getenv("EMBEDDER_OUT"))
    total = len(categories)

    vset_key = os.getenv("CATEGORIES_VSET_KEY", "categories:index")

    logging.info("migrating %d categories into vector set '%s'", total, vset_key)

    batch_size = 1000
    for i, row in enumerate(categories, start=1):
        cat_id = str(row["category_id"])
        embedding = row["embedding"]

        r.vset().vadd(vset_key, embedding, cat_id)

        attrs = {
            "name": row.get("category_name", ""),
            "label": row.get("category_label", ""),
            "level": row.get("category_level", 0),
        }
        r.vset().vsetattr(vset_key, cat_id, attrs)
        if i % batch_size == 0 or i == total:
            logging.info("migrated %d / %d", i, total)

    logging.info("done")


if __name__ == "__main__":
    migrate()
