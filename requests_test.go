package main

import "testing"
import "fmt"
import "net/http"
import "net/http/httptest"

func TestRequestPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	if code, _ := RequestPage(ts.URL); code != 404 {
		t.Errorf("Expected to return 404, got %d instead", code)
	}
}

func TestRequestSitemap(t *testing.T) {
	urls := []string{
		"http://google.com/maps",
		"http://google.com/docs",
	}
	xml := fmt.Sprintf(`
    <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
      <url>
        <loc>%v</loc>
        <lastmod>2016-04-04T02:08:53+03:00</lastmod>
        <priority>1.000000</priority>
      </url>
      <url>
        <loc>%v</loc>
        <lastmod>2016-04-04T01:12:13+03:00</lastmod>
        <priority>1.000000</priority>
      </url>
    </urlset>
  `, urls[0], urls[1])

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, xml)
	}))
	defer ts.Close()

	sitemap, _ := RequestSitemap(ts.URL)

	for i, url := range urls {
		if url != sitemap.URLS[i].Loc {
			t.Errorf("Expected %v, got %d instead", url, sitemap.URLS[i].Loc)
		}
	}
}
