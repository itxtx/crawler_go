package extractor

import (
	"fmt"
	"net/url"
)

// ExtractionError represents different types of extraction errors
type ExtractionError struct {
	Type    ErrorType
	Message string
	Source  string // The source that caused the error (URL, HTML snippet, etc.)
	Err     error  // Underlying error
}

// ErrorType represents different categories of extraction errors
type ErrorType string

const (
	ErrorTypeMalformedURL     ErrorType = "malformed_url"
	ErrorTypeMissingSrc       ErrorType = "missing_src"
	ErrorTypeJSONLDParse      ErrorType = "jsonld_parse_error"
	ErrorTypeHTMLParse        ErrorType = "html_parse_error"
	ErrorTypeXPathParse       ErrorType = "xpath_parse_error"
	ErrorTypeRegexParse       ErrorType = "regex_parse_error"
	ErrorTypeInvalidSelector  ErrorType = "invalid_selector"
	ErrorTypeNetworkError     ErrorType = "network_error"
	ErrorTypeUnknownPlatform  ErrorType = "unknown_platform"
	ErrorTypeEmptyContent     ErrorType = "empty_content"
)

// Error implements the error interface
func (e *ExtractionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (source: %s) - %v", e.Type, e.Message, e.Source, e.Err)
	}
	return fmt.Sprintf("%s: %s (source: %s)", e.Type, e.Message, e.Source)
}

// Unwrap allows error unwrapping for Go 1.13+ error handling
func (e *ExtractionError) Unwrap() error {
	return e.Err
}

// Is allows error comparison for Go 1.13+ error handling
func (e *ExtractionError) Is(target error) bool {
	if targetErr, ok := target.(*ExtractionError); ok {
		return e.Type == targetErr.Type
	}
	return false
}

// NewExtractionError creates a new ExtractionError
func NewExtractionError(errType ErrorType, message, source string, err error) *ExtractionError {
	return &ExtractionError{
		Type:    errType,
		Message: message,
		Source:  source,
		Err:     err,
	}
}

// NewMalformedURLError creates an error for malformed URLs
func NewMalformedURLError(rawURL string, err error) *ExtractionError {
	return NewExtractionError(
		ErrorTypeMalformedURL,
		"Failed to parse URL",
		rawURL,
		err,
	)
}

// NewMissingSrcError creates an error for missing src attributes
func NewMissingSrcError(elementType, context string) *ExtractionError {
	return NewExtractionError(
		ErrorTypeMissingSrc,
		fmt.Sprintf("Missing src attribute in %s element", elementType),
		context,
		nil,
	)
}

// NewJSONLDParseError creates an error for JSON-LD parsing failures
func NewJSONLDParseError(jsonContent string, err error) *ExtractionError {
	// Truncate long JSON content for cleaner error messages
	truncated := jsonContent
	if len(truncated) > 200 {
		truncated = truncated[:200] + "..."
	}
	return NewExtractionError(
		ErrorTypeJSONLDParse,
		"Failed to parse JSON-LD content",
		truncated,
		err,
	)
}

// NewHTMLParseError creates an error for HTML parsing failures
func NewHTMLParseError(htmlSnippet string, err error) *ExtractionError {
	// Truncate long HTML content for cleaner error messages
	truncated := htmlSnippet
	if len(truncated) > 200 {
		truncated = truncated[:200] + "..."
	}
	return NewExtractionError(
		ErrorTypeHTMLParse,
		"Failed to parse HTML content",
		truncated,
		err,
	)
}

// NewXPathParseError creates an error for XPath parsing failures
func NewXPathParseError(xpath string, err error) *ExtractionError {
	return NewExtractionError(
		ErrorTypeXPathParse,
		"Failed to parse XPath expression",
		xpath,
		err,
	)
}

// NewRegexParseError creates an error for regex parsing failures
func NewRegexParseError(pattern string, err error) *ExtractionError {
	return NewExtractionError(
		ErrorTypeRegexParse,
		"Failed to compile regex pattern",
		pattern,
		err,
	)
}

// ValidateURL validates and normalizes a URL, returning a custom error if invalid
func ValidateURL(rawURL string, base *url.URL) (string, error) {
	if rawURL == "" {
		return "", NewMissingSrcError("URL", "empty URL provided")
	}

	// Attempt to parse the URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", NewMalformedURLError(rawURL, err)
	}

	// If base URL is provided, resolve relative URLs
	if base != nil {
		resolved := base.ResolveReference(parsed)
		return resolved.String(), nil
	}

	// If no base URL, ensure the URL is absolute
	if !parsed.IsAbs() {
		return "", NewMalformedURLError(rawURL, fmt.Errorf("relative URL without base URL"))
	}

	return parsed.String(), nil
}

// SafeURLNormalize safely normalizes a URL with error handling
func SafeURLNormalize(rawURL string, base *url.URL) (string, *ExtractionError) {
	normalized, err := ValidateURL(rawURL, base)
	if err != nil {
		if extractionErr, ok := err.(*ExtractionError); ok {
			return "", extractionErr
		}
		return "", NewMalformedURLError(rawURL, err)
	}
	return normalized, nil
}
