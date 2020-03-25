# Heartbeat

![](https://travis-ci.org/luchkonikita/heartbeat.svg?branch=master)

This is a simple utility that allows you to check web-pages specified in a sitemap.
The program concurrently requests all the pages and prints the report with status
codes of received responses. Might be useful as a simple testing tool to find broken webpages
by looking for 5xx responses.

## Installation:

#### Option 1

- Download a binary from [releases](https://github.com/luchkonikita/heartbeat/releases) section.

#### Option 2
- Clone this repo into your $GOPATH.
- Build the binary with `go build`

## Example:

```
./heartbeat -limit=100 -concurrency=8 -timeout=500 -headers="one:1,two:2" -query='foo:bar' http://some-website.com/sitemap.xml
```


## Options:

```
  -concurrency int
      concurrency (default 5)
  -limit int
      limit for URLs to be checked (default 1000)
  -headers string
      headers to be send with requests
      string in format "headerName: headerValue,header2Name: header2Value"
  -timeout int
      timeout for requests (default 300)
```

## TODO

- [x] Write tests
- [x] Setup CI
- [x] Add reporter
- [x] Add HTTP basic auth support
- [ ] Add option to write report to a file
