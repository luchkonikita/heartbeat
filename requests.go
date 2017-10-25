package main

import "encoding/xml"
import "io/ioutil"
import "net/http"
import "time"

const (
	clientTimeout = time.Duration(25 * time.Second)
)

// newClient - Creates a pre-configured client
func newClient() http.Client {
	return http.Client{}
}

// requestPage - Requests a page by URL
func requestPage(c http.Client, url string, authParams AuthParams) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(authParams.Name, authParams.Password)
	if err != nil {
		return 0, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, err
}

// requestSitemap - Downloads and parses sitemap
func requestSitemap(c http.Client, url string, authParams AuthParams) (Sitemap, error) {
	sitemap := Sitemap{}

	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(authParams.Name, authParams.Password)
	if err != nil {
		return sitemap, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return sitemap, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sitemap, err
	}
	xml.Unmarshal(content, &sitemap)
	return sitemap, err
}
