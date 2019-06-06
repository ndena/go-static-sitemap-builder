package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"../link/linkparser"
)

var flagURL string
var flagTemplate string
var flagMaxDepth int

var hashURL = make(map[string]bool)
var visitedURL = make(map[string]bool)

type visitURL struct {
	URL   string
	Depth int
}

func init() {
	flag.StringVar(&flagURL, "url", "https://www.google.com/", "url to generate sitemap for")
	flag.StringVar(&flagTemplate, "template", "./tmpl/sitemap.xml", "path to sitemap template")
	flag.IntVar(&flagMaxDepth, "depth", 2, "maximum depth for sitemap look up")
	flag.Parse()
}

func main() {
	tmpl := template.Must(template.ParseFiles(flagTemplate))

	var links []string
	var urlsToVisit []visitURL

	urlsToVisit = append(urlsToVisit, visitURL{flagURL, 1})

	i := 0
	for i < len(urlsToVisit) {
		fmt.Printf("Visiting %v with depth %d\n", urlsToVisit[i].URL, urlsToVisit[i].Depth)
		newLinks := filterLinks(getLinks(urlsToVisit[i].URL))
		visitedURL[urlsToVisit[i].URL] = true

		if len(newLinks) > 0 {
			links = append(links, newLinks...)
			if urlsToVisit[i].Depth < flagMaxDepth {
				for _, l := range newLinks {
					if _, ok := visitedURL[l]; !ok {
						urlsToVisit = append(urlsToVisit, visitURL{l, urlsToVisit[i].Depth + 1})
					}
				}
			}
		}
		i++
	}

	tmpl.Execute(os.Stdout, struct {
		Links []string
	}{
		Links: links,
	})

	fmt.Println()
}

func getLinks(f string) []linkparser.Link {
	resp, err := http.Get(f)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return linkparser.ParseATags(body)
}

func filterLinks(links []linkparser.Link) []string {
	var result []string

	for _, l := range links {
		if len(l.Href) == 0 {
			continue
		}
		if l.Href[0] == '/' {
			l.Href = flagURL + l.Href[1:]
			if _, ok := hashURL[l.Href]; !ok {
				result = append(result, l.Href)
				hashURL[l.Href] = true
			}
		} else if len(l.Href) >= len(flagURL) && l.Href[:len(flagURL)] == flagURL {
			if _, ok := hashURL[l.Href]; !ok {
				result = append(result, l.Href)
				hashURL[l.Href] = true
			}
		}
	}
	return result
}
