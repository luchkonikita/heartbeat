![](https://travis-ci.org/luchkonikita/heartbeat.svg?branch=master)

# Heartbeat

This is a simple utility that allows you to check web-pages specified in a sitemap.
The program concurrently requests all the pages and prints the report with status
codes of received responses. Might be useful as a simple testing tool to find broken webpages
by looking for 5xx responses.

Example:

```
./heartbeat -limit=100 -concurrency=8 -timeout=500 http://some-website.com/sitemap.xml
```


Options:

```
  -concurrency int
      concurrency (default 5)
  -limit int
      limit for URLs to be checked (default 1000)
  -timeout int
      timeout for requests (default 300)
```

## TODO

- [x] Write tests
- [x] Setup CI
- [x] Add reporter
- [ ] Add HTTP basic auth support
- [ ] Add option to write report to a file