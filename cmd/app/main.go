package main

import (
	"os"

	"lbc/internal/app"
)

const (
	urlHome     string = "https://tocandraw.com/"
	urlPost     string = "https://tocandraw.com/post-sitemap.xml"
	urlPage     string = "https://tocandraw.com/page-sitemap.xml"
	urlCategory string = "https://tocandraw.com/category-sitemap.xml"
	urlTag      string = "https://tocandraw.com/post_tag-sitemap.xml"
	urlAuthor   string = "https://tocandraw.com/author-sitemap.xml"
)

func main() {
	app.ReadCommand()

	allSitemap := []string{urlPost, urlPage, urlCategory, urlTag, urlAuthor}
	crawler := app.NewCrawler(allSitemap, false, urlHome)
	if err := crawler.PurgeCache(); err != nil {
		os.Exit(1)
	}

	crawler.Run()

	mobileCrawler := app.NewCrawler(allSitemap, true, urlHome)
	mobileCrawler.Run()
}
