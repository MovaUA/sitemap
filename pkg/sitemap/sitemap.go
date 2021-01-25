package sitemap

import (
	"encoding/xml"

	_ "github.com/movaua/link/pkg/link"
)

// URLSet is a top-level model for sitemap.
// It encapsulates the file and references the current protocol standard.
// See https://www.sitemaps.org/protocol.html
type URLSet struct {
	XMLName xml.Name   `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Entries []URLEntry `xml:"url"`
}

// URLEntry is a parent tag for each URL entry
type URLEntry struct {
	URL string `xml:"loc"`
	Err error  `xml:"-"`
}

// Build walks all the links on the site provided by url
// and returns sitemap with links belonging to the same domain
// as of the provided url, or an error
func Build(url string) (URLSet, error) {
	// TODO: implement building of URLSet
	return URLSet{}, nil
}
