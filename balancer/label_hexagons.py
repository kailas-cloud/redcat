import csv
import h3
import sys

CSV_PATH = ".data/cyprus/pois.csv"
OUT_PATH = ".data/cyprus/pois_with_h3.csv"

def main():
    with open(CSV_PATH, "r", encoding="utf-8") as f_in, \
            open(OUT_PATH, "w", encoding="utf-8", newline="") as f_out:

        reader = csv.DictReader(f_in)
        fieldnames = reader.fieldnames + ["hex_5", "hex_6", "hex_7", "hex_8"]
        writer = csv.DictWriter(f_out, fieldnames=fieldnames)
        writer.writeheader()

        for row in reader:
            try:
                lat = float(row["latitude"])
                lng = float(row["longitude"])
            except Exception:
                continue

            row["hex_5"] = h3.latlng_to_cell(lat, lng, 5)
            row["hex_6"] = h3.latlng_to_cell(lat, lng, 6)
            row["hex_7"] = h3.latlng_to_cell(lat, lng, 7)
            row["hex_8"] = h3.latlng_to_cell(lat, lng, 8)

            writer.writerow(row)

    print("Готово. Файл сохранен в", OUT_PATH)


if __name__ == "__main__":
    main()
