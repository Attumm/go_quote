import csv

def custom_quoting(field):
    return field.startswith('"') and field.endswith('"')


def parser(item):
    item["author"] = item["author"].strip()
    if "  " in item["author"]:
        item["author"] = ' '.join(item['author'].split())

    if "," in item["author"]:
        item["author"] = item["author"].split(",")[0]

    item["quote"] = item["quote"].strip().capitalize()
    item["author"] = item["author"].strip()

    return item


invalid_tags = {"kink", "bdsm", "erotic", "sex"}


def validate(item, invalid_tags=set()):
    if not item["author"] or not item["quote"]:
        return False
    for tag in invalid_tags:
        if tag in item["category"]:
            #print("flagged: invalid_tags", item)
            return False

    if item["author"].count(" ") > 4:
        #print("flagged: too many spaces", item)
        return False
    return True


def main(base_path: str, in_file: str, out_file: str):
    with open(base_path + in_file, 'r', newline='', encoding='utf-8') as input_file, \
         open(base_path + out_file, 'w', newline='', encoding='utf-8') as output_file:

        reader = csv.DictReader(input_file)
        fieldnames = reader.fieldnames

        writer = csv.DictWriter(output_file, fieldnames=fieldnames,
                                quoting=csv.QUOTE_MINIMAL,
                                quotechar='"')

        writer.writeheader()

        successful = 0
        total = 0
        for row in reader:
            total += 1
            
            parsed_item = parser(row)

            if not validate(parsed_item, invalid_tags):
                #print(parsed_item)
                continue

            quoted_row = {k: (v if custom_quoting(v) else v.replace('"', '""')) for k, v in parsed_item.items()}
            writer.writerow(quoted_row)
            successful += 1

        print(f"csv added rows: {successful} total: {total}")

if __name__ == "__main__":
    base_path = "data/"
    in_file = "quotes.csv"
    out_file = "filtered.csv"
    main(base_path, in_file, out_file)
