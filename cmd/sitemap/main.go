package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/movaua/sitemap/pkg/sitemap"
)

func main() {
	url := flag.String("url", "", "[required] an URL to the site")
	flag.Parse()
	if *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	urlset, err := sitemap.Build(*url)
	if err != nil {
		log.Fatalln(err)
	}

	out, err := xml.MarshalIndent(urlset, "", "  ")

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%s%s\n", xml.Header, out)
}
