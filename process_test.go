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
		if r.Header.Get("Auth") != "Yes" {
			w.WriteHeader(403)
			return
		}

		query := r.URL.Query()

		if len(query["page"]) > 0 {
			pageNum := query["page"][0]
			if pageNum == "1" {
				w.WriteHeader(200)
			}

			if pageNum == "2" || pageNum == "3" {
				w.WriteHeader(500)
			}

			return
		}

		if len(query["sitemap"]) > 0 {
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
					<url>
						<loc>http://127.0.0.1:8080?page=3</loc>
						<lastmod>2016-04-04T01:12:13+03:00</lastmod>
						<priority>1.000000</priority>
					</url>
        </urlset>
      `
			fmt.Fprintln(w, xml)
		} else if len(query["foo"]) == 0 {
			t.Error("Expected to assign a custom query string")
		}
	}))

	ts.Listener.Close()
	ts.Listener = l

	defer ts.Close()

	ts.Start()

	buf := new(bytes.Buffer)
	headers := parameters{parameter{name: "Auth", value: "Yes"}}
	query := parameters{parameter{name: "foo", value: "bar"}}

	success := process(buf, 1, 10, 200, "http://127.0.0.1:8080?sitemap=true", headers, query)

	if success {
		t.Error("Expected to return false")
	}

	expected := `+----+------------------------------+--------+
| NO |             URL              | STATUS |
+----+------------------------------+--------+
|  1 | http://127.0.0.1:8080?page=2 |    500 |
|  2 | http://127.0.0.1:8080?page=3 |    500 |
+----+------------------------------+--------+
`

	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}
}
