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

	"lbc/internal/entity"
	"lbc/pkg/log"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

type Crawler struct {
	isMobile   bool
	siteMapURL []string
	urlInSite  []urlInSite
	additional []string

	urlMap     map[string]struct{}
	urlMapLock sync.RWMutex

	client *resty.Client
	logger *log.Log
}

type urlInSite struct {
	urlArr []string
}

func NewCrawler(siteMapURL []string, isMobile bool, additional ...string) *Crawler {
	c := &Crawler{
		siteMapURL: siteMapURL,
		isMobile:   isMobile,
		client:     resty.New(),
		logger:     log.Get(),
		urlMap:     make(map[string]struct{}),
		additional: additional,
	}

	if isMobile {
		c.client.SetHeader("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	}

	if err := c.getSite(); err != nil {
		c.logger.Error(err)
		return nil
	}

	return c
}

func (c *Crawler) getSite() error {
	c.urlInSite = append(c.urlInSite, urlInSite{urlArr: c.additional})
	for _, url := range c.siteMapURL {
		resp, err := c.client.R().Get(url)
		if err != nil {
			return err
		}

		var urlset entity.SiteMap
		err = xml.Unmarshal(resp.Body(), &urlset)
		if err != nil {
			return err
		}

		var data urlInSite
		for _, url := range urlset.URL {
			data.urlArr = append(data.urlArr, url.Loc)
		}
		c.urlInSite = append(c.urlInSite, data)
		c.logger.Infof("Found %d urls in %s", len(data.urlArr), url)
	}
	return nil
}

func (c *Crawler) getHTMLBody(url string) ([]byte, error) {
	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		c.logger.Errorf("404: %s", url)
		return nil, nil
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("status code not ok")
	}

	return resp.Body(), nil
}

func (c *Crawler) foundURL(n *html.Node, urlMap map[string]struct{}) {
	for _, attr := range n.Attr {
		for _, split := range strings.Split(attr.Val, " ") {
			parseURL, err := url.Parse(split)
			if err != nil {
				continue
			} else if !parseURL.IsAbs() {
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
				if c.appendURL(parseURL.String()) {
					urlMap[parseURL.String()] = struct{}{}
				}
			}
		}
	}
	for node := n.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode {
			c.foundURL(node, urlMap)
		}
	}
}

func (c *Crawler) appendURL(url string) bool {
	c.urlMapLock.RLock()
	_, ok := c.urlMap[url]
	c.urlMapLock.RUnlock()
	if ok {
		return false
	}
	c.urlMapLock.Lock()
	c.urlMap[url] = struct{}{}
	c.urlMapLock.Unlock()
	return true
}

func (c *Crawler) crawl(url string, waitGroup *sync.WaitGroup) {
	c.logger.Infof("Crawling %s", url)
	if waitGroup != nil {
		defer waitGroup.Done()
	}
	bo, err := c.getHTMLBody(url)
	if err != nil {
		c.logger.Error(err)
		return
	}

	reader := bytes.NewReader(bo)
	doc, err := html.Parse(reader)
	if err != nil {
		c.logger.Error(err)
		return
	}

	urlMap := make(map[string]struct{})
	c.foundURL(doc, urlMap)
	for k := range urlMap {
		c.crawl(k, nil)
	}
}

func (c *Crawler) Run() {
	waitGroup := &sync.WaitGroup{}
	for _, v := range c.urlInSite {
		var concurrency int
		for _, url := range v.urlArr {
			if concurrency > 10 {
				waitGroup.Wait()
				concurrency = 0
			}
			concurrency++
			waitGroup.Add(1)
			go c.crawl(url, waitGroup)
		}
		waitGroup.Wait()
	}
}
