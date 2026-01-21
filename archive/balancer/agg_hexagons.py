import csv
import h3
from collections import defaultdict

CSV_PATH = ".data/cyprus/pois.csv"

OUT_FILES = {
    5: ".data/cyprus/hex_5.csv",
    6: ".data/cyprus/hex_6.csv",
    7: ".data/cyprus/hex_7.csv",
    8: ".data/cyprus/hex_8.csv",
}


def main():
    counters = {
        res: defaultdict(int)
        for res in OUT_FILES.keys()
    }

    with open(CSV_PATH, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)

        for row in reader:
            try:
                lat = float(row["latitude"])
                lng = float(row["longitude"])
            except Exception:
                continue

            for res in counters.keys():
                h = h3.latlng_to_cell(lat, lng, res)
                counters[res][h] += 1

    for res, out_path in OUT_FILES.items():
        with open(out_path, "w", encoding="utf-8", newline="") as f_out:
            writer = csv.writer(f_out)
            writer.writerow(["hex_id", "places_count"])

            for hex_id, count in counters[res].items():
                writer.writerow([hex_id, count])

        print(f"Записано: {out_path}")


if __name__ == "__main__":
    main()
