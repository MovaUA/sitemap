package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/movaua/sitemap/pkg/sitemap"
)

func main() {
	urlFlag := flag.String("url", "https://github.com/movaua/sitemap", "the URL of the site you want to build a sitemap for")
	maxDepth := flag.Int("depth", 3, "a maximum depth of pages to traverse")
	timeout := flag.Int("timeout", 3, "timeout in seconds to wait for response from a single HTTP request")
	flag.Parse()

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	builder := sitemap.NewBuilder(sitemap.WithClient(client))

	urlset, err := builder.Build(*urlFlag, *maxDepth)
	if err != nil {
		log.Fatalln(err)
	}

	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")

	fmt.Print(xml.Header)
	if err = enc.Encode(urlset); err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
}
