package main

import "testing"
import "bytes"
import "fmt"
import "net"
import "net/http"
import "net/http/httptest"

func TestProcess(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()["sitemap"]) != 0 {
			xml := `
        <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
          <url>
            <loc>http://127.0.0.1:8080?page=1</loc>
            <lastmod>2016-04-04T02:08:53+03:00</lastmod>
            <priority>1.000000</priority>
          </url>
          <url>
            <loc>http://127.0.0.1:8080?page=2</loc>
            <lastmod>2016-04-04T01:12:13+03:00</lastmod>
            <priority>1.000000</priority>
          </url>
        </urlset>
      `
			fmt.Fprintln(w, xml)
		} else {
			w.WriteHeader(200)
		}
	}))

	ts.Listener.Close()
	ts.Listener = l

	defer ts.Close()

	ts.Start()

	buf := new(bytes.Buffer)
	process(buf, 8, 10, 200, "http://127.0.0.1:8080?sitemap=true")

	expected := `+----+------------------------------+--------+
| NO |             URL              | STATUS |
+----+------------------------------+--------+
|  1 | http://127.0.0.1:8080?page=1 |    200 |
|  2 | http://127.0.0.1:8080?page=2 |    200 |
+----+------------------------------+--------+
`

	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}
}
