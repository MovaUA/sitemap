// Package sitemap builds sitemap for the site.
// Built sitemap is suitable for XML encoding in Sitemaps XML format.
// See https://www.sitemaps.org/protocol.html
package sitemap

// Build walks all the links on the site provided by url
// and returns sitemap with links belonging to the same domain
// as of the provided url, or an error if any.
func Build(rootURL string, opts ...OptionFunc) (*URLSet, error) {
	return NewBuilder(opts...).Build(rootURL)
}
