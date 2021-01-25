package sitemap

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/movaua/link/pkg/link"
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
func Build(rawurl string) (*URLSet, error) {
	startURL, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	urls := make(map[string]struct{})

	foundCh := make(chan []*url.URL)
	findCh := make(chan *url.URL, 4)

	findCh <- startURL

	timeout := time.NewTimer(3 * time.Second)

tasks:
	for {
		select {
		case queuedURL := <-findCh:
			go func(u *url.URL) {
				urls, err := findURLs(u)
				if err != nil {
					log.Printf("%s: %v\n", u, err)
					return
				}
				foundCh <- urls
			}(queuedURL)
		case foundURLs := <-foundCh:
			for _, foundURL := range foundURLs {
				rawurl := foundURL.String()
				_, seen := urls[rawurl]
				if ok := !seen && foundURL.Host == startURL.Host; !ok {
					continue
				}

				urls[rawurl] = struct{}{}
				findCh <- foundURL
			}
		case <-timeout.C:
			break tasks
		}
	}

	entries := make([]URLEntry, 0, len(urls))
	for rawurl, _ := range urls {
		entries = append(entries, URLEntry{URL: rawurl})
	}

	return &URLSet{Entries: entries}, nil
}

func findURLs(u *url.URL) ([]*url.URL, error) {
	rawurl := u.String()
	res, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got %d %s", res.StatusCode, res.Status)
	}

	links, err := link.Find(res.Body)
	if err != nil {
		return nil, err
	}

	urls := make([]*url.URL, 0, len(links))
	for _, link := range links {
		parsedURL, err := url.Parse(link.Href)
		if err != nil {
			return nil, err
		}
		parsedURL = u.ResolveReference(parsedURL)
		urls = append(urls, parsedURL)
	}

	return urls, nil
}
