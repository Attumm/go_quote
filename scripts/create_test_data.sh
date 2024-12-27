echo "started collecting testing data"
mkdir -p test_data
rm test_data/*

#curl -s "http://127.0.0.1:8000/quotes?page_size=0&format=json" > test_data/expected_page_size_0.json &
curl -s "http://127.0.0.1:8000/quotes?page_size=1&format=json" > test_data/expected_page_size_1.json &
curl -s "http://127.0.0.1:8000/quotes?page_size=2&format=json" > test_data/expected_page_size_2.json &
curl -s "http://127.0.0.1:8000/quotes?page_size=3&format=json" > test_data/expected_page_size_3.json &
curl -s "http://127.0.0.1:8000/quotes?page_size=10&format=json" > test_data/expected_page_size_10.json &
echo "done collecting testing data"
