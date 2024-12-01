# Go Quote

A performance focussed API for serving random quotes with each API call.
The API is backed by an in-memory database containing close to **500,000 quotes**, capable of handling high traffic efficiently 
load test achieving **32k+ requests per second**.


---

## Features

- **Massive Database**: 500k quotes loaded in-memory for instant access.
- **High Performance**: Tested with Apache Bench at 32k requests/second.
- **Multiple Formats**: Returns quotes in a variety of output formats (see below).

---

## Documentation

Access the API documentation here:
- Local: [`http://127.0.0.1:8000/docs/`](http://127.0.0.1:8000/docs/)
- Live: [`https://go-quote.azurewebsites.net/docs`](https://go-quote.azurewebsites.net/docs)

---

## API Live URL

The API is running live at:  
[`https://go-quote.azurewebsites.net/`](https://go-quote.azurewebsites.net/)

---

## Supported Formats

The API supports the following output formats:

| Format       | Content-Type                      |  
|--------------|-----------------------------------|  
| **xml**      | `application/xml`                |  
| **html**     | `text/html`                      |  
| **json**     | `application/json`               |  
| **text**     | `text/plain`                     |  
| **markdown** | `text/markdown`                  |  
| **yaml**     | `application/yaml`               |  
| **csv**      | `text/csv`                       |  
| **rss**      | `application/rss+xml`            |  
| **atom**     | `application/atom+xml`           |  
| **oembed**   | `application/json+oembed`        |  
| **oembed.xml** | `text/xml+oembed`              |  
| **embed**    | `text/html`                      |  
| **embed.js** | `application/javascript`         |
---

## Quick Start with Docker

Run the API locally using Docker:
```bash
docker run --rm -it -p 8000:8000 $(docker build -q .)
```

---

## Performance Testing

The API has been load-tested using **Apache Bench**
Getting close to 35k requests per second.
Results from a macpro m2
```bash
ab -n 10000 -c 100 -l http://127.0.0.1:8000/
```

```bash
$ ab -n 10000 -c 100 -l http://127.0.0.1:8000/
This is ApacheBench, Version 2.3 <$Revision: 1903618 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:
Server Hostname:        127.0.0.1
Server Port:            8000

Document Path:          /
Document Length:        Variable

Concurrency Level:      100
Time taken for tests:   0.295 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      4441717 bytes
HTML transferred:       3352049 bytes
Requests per second:    33901.75 [#/sec] (mean)
Time per request:       2.950 [ms] (mean)
Time per request:       0.029 [ms] (mean, across all concurrent requests)
Transfer rate:          14705.27 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   0.7      1       6
Processing:     1    1   0.6      1       6
Waiting:        1    1   0.5      1       6
Total:          2    3   1.2      3      11

Percentage of the requests served within a certain time (ms)
  50%      3
  66%      3
  75%      3
  80%      3
  90%      3
  95%      4
  98%      8
  99%     10
 100%     11 (longest request)
```

---

### Local Development

- API Base URL: [`http://127.0.0.1:8000`](http://127.0.0.1:8000)
- Docs: [`http://127.0.0.1:8000/docs`](http://127.0.0.1:8000/docs)