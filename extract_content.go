package main

// import statements
import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
)

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

func extractContent(htmlContent, pattern, selectorType, format string, printContent bool) {
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

func extractMultipleContents(htmlContent string, patterns []string, selectorType string, format string, printContent bool) {
	for _, pattern := range patterns {
		extractContent(htmlContent, pattern, selectorType, format, printContent)
	}
}
