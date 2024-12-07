 
python3 scripts/filter_data/main.py 
rm data/quotes.bytesz

./scripts/convert_data/convert_data --FILENAME data/filtered.csv 

mv filtered.bytesz data/quotes.bytesz

