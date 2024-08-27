package extractor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
)

type LinkInfo struct {
	URL         string
	Description string
}

func formatOutput(matches []string, format string) string {
	switch format {
	case "json":
		jsonData, _ := json.Marshal(matches)
		return string(jsonData)
	case "csv":
		return strings.Join(matches, ",")
	default:
		return strings.Join(matches, "\n")
	}
}

func ExtractContent(htmlContent, pattern, selectorType, format string, printContent bool) {
	var matches []string
	switch selectorType {
	case "css":
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}
		doc.Find(pattern).Each(func(i int, s *goquery.Selection) {
			matches = append(matches, s.Text())
		})
	case "xpath":
		doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}
		expr := xpath.MustCompile(pattern)
		nodes := htmlquery.Find(doc, expr.String())
		for _, node := range nodes {
			matches = append(matches, htmlquery.InnerText(node))
		}
	case "regex":
		re := regexp.MustCompile(pattern)
		reMatches := re.FindAllStringSubmatch(htmlContent, -1)
		for _, match := range reMatches {
			if len(match) > 1 {
				matches = append(matches, match[1])
			}
		}
	}

	if len(matches) == 0 {
		fmt.Println("No content matched the pattern.")
		return
	}

	output := formatOutput(matches, format)
	fmt.Println(output)

	fmt.Println("Extracted content:")
	for _, match := range matches {
		fmt.Println(match)
	}
}

func ExtractMultipleContents(htmlContent string, patterns []string, selectorType string, format string, printContent bool) {
	for _, pattern := range patterns {
		ExtractContent(htmlContent, pattern, selectorType, format, printContent)
	}
}

func ExtractLinksAndDescriptions(htmlBody string, baseURL *url.URL, filter string) ([]LinkInfo, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	var links []LinkInfo
	var extractFunc func(*html.Node)
	extractFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href, text string
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}
			if href != "" {
				absoluteURL := baseURL.ResolveReference(&url.URL{Path: href}).String()
				if filter == "" || strings.Contains(absoluteURL, filter) {
					text = extractText(n)
					links = append(links, LinkInfo{URL: absoluteURL, Description: text})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractFunc(c)
		}
	}
	extractFunc(doc)
	return links, nil
}

func extractText(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode {
			text += extractText(c)
		}
	}
	return strings.TrimSpace(text)
}
