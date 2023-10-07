// Package app package app
package app

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"lbc/internal/entity"
	"lbc/pkg/log"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

const (
	urlHome     string = "https://tocandraw.com/"
	urlPost     string = "https://tocandraw.com/post-sitemap.xml"
	urlPage     string = "https://tocandraw.com/page-sitemap.xml"
	urlCategory string = "https://tocandraw.com/category-sitemap.xml"
	urlTag      string = "https://tocandraw.com/post_tag-sitemap.xml"
	urlAuthor   string = "https://tocandraw.com/author-sitemap.xml"
)

var (
	clientSingle *resty.Client
	urlMapLock   sync.RWMutex
	urlMap       = make(map[string]bool)
	logger       = log.Get()
	isMobile     = false
)

func Run() {
	allSitemap := []string{urlPost, urlPage, urlCategory, urlTag, urlAuthor}
	allURL := []string{}
	for _, url := range allSitemap {
		urls, err := getSite(url)
		if err != nil {
			logger.Error(err)
			continue
		}
		allURL = append(allURL, urls...)
	}

	allURL = append(allURL, urlHome)
	logger.Info("All URL: ", len(allURL))
	for _, url := range allURL {
		tmpURL := url
		go func() {
			err := crawl(tmpURL)
			if err != nil {
				logger.Error(err)
				return
			}
		}()
	}
}

func Clear() {
	urlMapLock.Lock()
	urlMap = make(map[string]bool)
	urlMapLock.Unlock()
	clientSingle = nil
}

func SetMobile() {
	isMobile = true
}

func getSite(url string) ([]string, error) {
	client := getHTTPClient()
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var urlset entity.SiteMap
	err = xml.Unmarshal(resp.Body(), &urlset)
	if err != nil {
		return nil, err
	}
	var urls []string
	for _, url := range urlset.URL {
		urls = append(urls, url.Loc)
	}
	return urls, nil
}

func getHTTPClient() *resty.Client {
	if clientSingle != nil {
		return clientSingle
	}

	newClient := resty.New()
	if isMobile {
		newClient.SetHeader("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	}
	clientSingle = newClient
	return clientSingle
}

func crawl(crawlURL string) error {
	urlMapLock.RLock()
	_, ok := urlMap[crawlURL]
	urlMapLock.RUnlock()
	if ok {
		return nil
	}

	urlMapLock.Lock()
	urlMap[crawlURL] = true
	urlMapLock.Unlock()

	client := getHTTPClient()
	time.Sleep(1500 * time.Millisecond)
	resp, err := client.R().Get(crawlURL)
	if err != nil {
		return err
	}

	p, err := url.QueryUnescape(crawlURL)
	if err == nil {
		logger.Info(p)
	}

	if resp.StatusCode() == http.StatusNotFound {
		logger.Errorf("404: %s", crawlURL)
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("status code not ok")
	}

	reader := bytes.NewReader(resp.Body())
	doc, err := html.Parse(reader)
	if err != nil {
		return err
	}

	extractURLs(doc)
	return nil
}

func extractURLs(n *html.Node) {
	defer func() {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractURLs(c)
		}
	}()

	if n.Type != html.ElementNode {
		return
	}

	var foundURL []string
	for _, attr := range n.Attr {
		splits := strings.Split(attr.Val, " ")
		for _, split := range splits {
			parseURL, err := url.Parse(split)
			if err != nil {
				continue
			}
			switch {
			case !strings.Contains(parseURL.String(), "tocandraw.com"):
				continue
			case strings.Contains(parseURL.String(), ".php"):
				continue
			case strings.Contains(parseURL.String(), "wp-json"):
				continue
			default:
				foundURL = append(foundURL, parseURL.String())
			}
		}
	}

	for _, v := range foundURL {
		go func(url string) {
			if err := crawl(url); err != nil {
				logger.Error(err)
				return
			}
		}(v)
	}
}
