package main

import "encoding/xml"
import "io/ioutil"
import "net/http"

// Requests a page by URL
func RequestPage(url string) (int, error) {
	resp, err := http.Get(url)
	return resp.StatusCode, err
}

// Downloads and parses sitemap
func RequestSitemap(url string) (Sitemap, error) {
	sitemap := Sitemap{}
	resp, err := http.Get(url)
	if err != nil {
		return sitemap, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sitemap, err
	}
	xml.Unmarshal(content, &sitemap)
	return sitemap, nil
}
