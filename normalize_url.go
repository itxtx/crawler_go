package main

import (
	"net/url"
	"strings"
)

// normalizeURL normalizes the given URL string by removing the scheme (http/https)
// and the "www" subdomain if present. It also ensures there's no trailing slash.
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
