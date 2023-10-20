// Package app package app
package app

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"lbc/internal/entity"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
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
}

type urlInSite struct {
	urlArr []string
}

func NewCrawler(siteMapURL []string, isMobile bool, additional ...string) *Crawler {
	c := &Crawler{
		siteMapURL: siteMapURL,
		isMobile:   isMobile,
		client:     resty.New(),
		urlMap:     make(map[string]struct{}),
		additional: additional,
	}

	if isMobile {
		c.client.SetHeader(
			"User-Agent",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		)
	}

	if err := c.getSite(); err != nil {
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
		fmt.Printf("Found %d urls in %s\n", len(data.urlArr), url)
	}
	return nil
}

func (c *Crawler) getHTMLBody(target string) ([]byte, error) {
	c.urlMapLock.RLock()
	_, ok := c.urlMap[target]
	c.urlMapLock.RUnlock()
	if ok {
		return nil, nil
	}

	c.urlMapLock.Lock()
	c.urlMap[target] = struct{}{}
	c.urlMapLock.Unlock()

	if c.isMobile {
		if strings.Contains(target, "?noamp=mobile") {
			target = strings.ReplaceAll(target, "?noamp=mobile", "?amp=1")
		}
	} else {
		if strings.Contains(target, "?amp=1") {
			target = strings.ReplaceAll(target, "?amp=1", "?noamp=mobile")
		}
	}

	// if un, err := url.PathUnescape(target); err == nil {
	// 	fmt.Println(un)
	// }
	fmt.Print("+")
	resp, err := c.client.R().Get(target)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("status code not ok")
	}

	return resp.Body(), nil
}

const (
	topDomain     = "tocandraw.com"
	urlTypePhp    = ".php"
	urlTypeWpJSON = "wp-json"
)

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
			case !strings.Contains(parseURL.String(), topDomain):
				continue
			case strings.Contains(parseURL.String(), urlTypePhp), strings.Contains(parseURL.String(), urlTypeWpJSON):
				continue
			default:
				if _, ok := urlMap[parseURL.String()]; !ok {
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

func (c *Crawler) crawl(url string, waitGroup *sync.WaitGroup) {
	if waitGroup != nil {
		defer waitGroup.Done()
	}

	bo, err := c.getHTMLBody(url)
	if err != nil {
		return
	}

	if bo == nil {
		return
	}

	reader := bytes.NewReader(bo)
	doc, err := html.Parse(reader)
	if err != nil {
		return
	}

	subWaitGroup := &sync.WaitGroup{}
	urlMap := make(map[string]struct{})
	c.foundURL(doc, urlMap)
	for k := range urlMap {
		subWaitGroup.Add(1)
		go c.crawl(k, subWaitGroup)
	}
	subWaitGroup.Wait()
}

func (c *Crawler) Run() {
	waitGroup := &sync.WaitGroup{}
	for _, v := range c.urlInSite {
		var concurrency int
		for _, url := range v.urlArr {
			if concurrency > 30 {
				waitGroup.Wait()
				concurrency = 0
			}
			concurrency++
			waitGroup.Add(1)
			if c.isMobile {
				url = fmt.Sprintf("%s?amp=1", url)
			}
			go c.crawl(url, waitGroup)
		}
		waitGroup.Wait()
	}
}

func (c *Crawler) PurgeCache() error {
	cloudflareAuth := viper.GetString("cloudflare-auth")
	cloudflareZoneID := viper.GetString("cloudflare-zone")

	resp, err := c.client.R().
		SetHeader("Authorization", cloudflareAuth).
		SetBody(map[string]any{
			"purge_everything": true,
		}).
		Post(fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", cloudflareZoneID))
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		fmt.Println(resp.String())
		return errors.New("status code not ok")
	}

	fmt.Println("Purge cache success, wait 45 second")
	time.Sleep(45 * time.Second)
	return nil
}
