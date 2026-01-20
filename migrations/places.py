import csv
import os
import logging
import redis

def get_redis():
    return redis.Redis(
        host=os.getenv("REDIS_HOST", "localhost"),
        port=int(os.getenv("REDIS_PORT", "6379")),
        db=int(os.getenv("REDIS_DB", "0")),
        password=os.getenv("REDIS_PASSWORD") or None,
        decode_responses=True,
    )

def migrate_pois(path: str):
    r = get_redis()

    logging.basicConfig(level=logging.INFO, format="%(message)s")
    logging.info(f"Loading POIs from {path}")

    with open(path, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        total = 0

        for row in reader:
            total += 1

            place_id = row["fsq_place_id"].strip()
            key = f"place:{place_id}"
            r.hset(key, mapping=row)

            if total % 2000 == 0:
                logging.info(f"Migrated {total}...")

        logging.info(f"Done. Total {total} POIs migrated.")

if __name__ == "__main__":
    path = os.getenv("POIS_CSV", ".data/cyprus/pois.csv")
    migrate_pois(path)
