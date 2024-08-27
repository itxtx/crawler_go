package crawl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func fetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("received error status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		return "", fmt.Errorf("content is not text/html")
	}

	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(htmlData), nil
}

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

func normalizeURL(inputURL string) (string, error) {
	u, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	// Remove scheme (http, https)
	normalized := u.Hostname() + u.EscapedPath()

	// Remove "www" subdomain if present
	normalized = strings.TrimPrefix(normalized, "www.")

	// Remove trailing slash if present
	normalized = strings.TrimSuffix(normalized, "/")

	// Append query parameters if they exist
	if u.RawQuery != "" {
		normalized += "?" + u.RawQuery
	}

	return normalized, nil
}
