// Package sitemap builds sitemap for the site.
// Built sitemap is suitable for XML encoding in Sitemaps XML format.
// See https://www.sitemaps.org/protocol.html
package sitemap

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/movaua/link/pkg/link"
)

// Builder buils sitemap
type Builder struct {
	client             *http.Client
	filter             FilterFunc
	maxDepth           int
	concurrentRequests int
}

// NewBuilder creates a builder
func NewBuilder(opts ...OptionFunc) *Builder {
	b := &Builder{
		client:             http.DefaultClient,
		filter:             defaultFilter,
		maxDepth:           3,
		concurrentRequests: runtime.NumCPU(),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Build walks all the links on the site provided by url
// and returns sitemap with links belonging to the same domain
// as of the provided url, or an error if any
func (b *Builder) Build(rootURL string) (*URLSet, error) {
	res, err := b.client.Get(rootURL)
	if err != nil {
		return nil, err
	}
	if err := res.Body.Close(); err != nil {
		return nil, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%s", res.Status)
	}

	root := res.Request.URL

	log.Debugf("concurrentRequests %d\n", b.concurrentRequests)

	var entries []URLEntry
	seen := make(map[string]struct{})

	var queue []string
	next := []string{
		root.String(),
	}

	for depth := 0; depth < b.maxDepth; depth++ {
		queue, next = next, nil

		log.Debugf("depth %d out of %d, queue length %d\n", depth+1, b.maxDepth, len(queue))

		jobs := make(chan string, b.concurrentRequests)
		results := make(chan []string, b.concurrentRequests)

		for i := 0; i < b.concurrentRequests; i++ {
			go func(jobs <-chan string, results chan<- []string) {
				for rawurl := range jobs {
					request, foundLinks, err := b.findLinks(rawurl)
					if err != nil {
						log.Warnf("could not find URLs from %q: %v\n", rawurl, err)
						results <- nil
						continue
					}

					result := make([]string, 0, len(foundLinks))
					for _, found := range foundLinks {
						filtered, ok := b.filter(root, request, found.Href)
						if !ok {
							continue
						}
						result = append(result, filtered.String())
					}

					results <- result
				}
			}(jobs, results)
		}

		go func() {
			defer close(jobs)
			for _, rawurl := range queue {
				jobs <- rawurl
			}
		}()

		for range queue {
			for _, rawurl := range <-results {
				if _, ok := seen[rawurl]; ok {
					continue
				}
				seen[rawurl] = struct{}{}
				next = append(next, rawurl)
			}
		}

		if len(next) == 0 {
			break
		}

		foundEntries := make([]URLEntry, 0, len(next))
		for _, rawurl := range next {
			foundEntries = append(foundEntries, URLEntry{URL: rawurl})
		}
		entries = append(entries, foundEntries...)
	}

	return &URLSet{Entries: entries}, nil
}

func (b *Builder) findLinks(rawurl string) (*url.URL, []link.Link, error) {
	res, err := b.client.Get(rawurl)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return res.Request.URL, nil, fmt.Errorf("%s", res.Status)
	}

	foundLinks, err := link.Find(res.Body)
	if err != nil {
		return res.Request.URL, nil, err
	}

	return res.Request.URL, foundLinks, nil
}

func defaultFilter(root, page *url.URL, pageLink string) (*url.URL, bool) {
	parsed, err := url.Parse(pageLink)
	if err != nil {
		return nil, false
	}
	if strings.HasPrefix(pageLink, "#") {
		return nil, false
	}

	resolved := page.ResolveReference(parsed)
	if ok := (resolved.Scheme == "http" || resolved.Scheme == "https") && resolved.Host == root.Host; !ok {
		return nil, false
	}

	return resolved, true
}
