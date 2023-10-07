// Package entity package entity
package entity

import "encoding/xml"

type SiteMap struct {
	XMLName        xml.Name `xml:"urlset"`
	Text           string   `xml:",chardata"`
	Xsi            string   `xml:"xsi,attr"`
	Image          string   `xml:"image,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Xmlns          string   `xml:"xmlns,attr"`
	URL            []struct {
		Text    string `xml:",chardata"`
		Loc     string `xml:"loc"`
		Lastmod string `xml:"lastmod"`
	} `xml:"url"`
}
