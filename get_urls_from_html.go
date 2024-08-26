package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// getURLsFromHTML extracts all URLs from the provided HTML body, converting relative URLs to absolute ones.
func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	base, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	var urls []string
	var extractLinks func(*html.Node)
	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link, err := base.Parse(attr.Val)
					if err == nil {
						urls = append(urls, link.String())
					}
					break
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			extractLinks(child)
		}
	}

	extractLinks(doc)
	// Return an empty slice instead of nil if no URLs were found
	if urls == nil {
		return []string{}, nil
	}

	return urls, nil
}
