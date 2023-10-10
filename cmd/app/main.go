package main

import "lbc/internal/app"

const (
	urlHome     string = "https://tocandraw.com/"
	urlPost     string = "https://tocandraw.com/post-sitemap.xml"
	urlPage     string = "https://tocandraw.com/page-sitemap.xml"
	urlCategory string = "https://tocandraw.com/category-sitemap.xml"
	urlTag      string = "https://tocandraw.com/post_tag-sitemap.xml"
	urlAuthor   string = "https://tocandraw.com/author-sitemap.xml"
)

func main() {
	allSitemap := []string{urlPost, urlPage, urlCategory, urlTag, urlAuthor}
	crawler := app.NewCrawler(allSitemap, false, urlHome)
	crawler.Run()

	mobileCrawler := app.NewCrawler(allSitemap, true, urlHome)
	mobileCrawler.Run()
}
