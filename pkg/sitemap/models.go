package sitemap

import (
	"encoding/xml"
)

// URLSet is a sitemap model.
// See Sitemaps XML format at https://www.sitemaps.org/protocol.html
type URLSet struct {
	XMLName xml.Name   `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Entries []URLEntry `xml:"url"`
}

// URLEntry is a parent tag for each URL entry
type URLEntry struct {
	// URL of the page.
	// This URL must begin with the protocol (such as http) and end with a trailing slash,
	// if your web server requires it. This value must be less than 2,048 characters.
	URL string `xml:"loc"`
}
