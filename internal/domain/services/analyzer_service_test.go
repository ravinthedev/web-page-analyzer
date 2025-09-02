package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLinkAccessibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`
				<!DOCTYPE html>
				<html>
				<head><title>Test Page</title></head>
				<body>
					<h1>Test Page</h1>
					<a href="/internal-page">Internal Link</a>
					<a href="relative-page">Relative Link</a>
					<a href="#fragment">Fragment Link</a>
					<a href="mailto:test@example.com">Email Link</a>
					<a href="tel:+1234567890">Phone Link</a>
				</body>
				</html>
			`))
		case "/internal-page":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>Internal Page</body></html>"))
		case "/relative-page":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>Relative Page</body></html>"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	wrappedClient := NewHTTPClient(http.DefaultClient)
	parser := NewHTMLParser(wrappedClient)
	service := NewAnalyzerService(wrappedClient, parser, 10)

	ctx := context.Background()
	result, err := service.AnalyzeURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	linkAnalysis := result.Links

	totalLinks := linkAnalysis.Internal + linkAnalysis.External
	if totalLinks != 5 {
		t.Errorf("Expected 5 total links, got %d", totalLinks)
	}

	if linkAnalysis.Internal != 3 {
		t.Errorf("Expected 3 internal links, got %d", linkAnalysis.Internal)
	}

	if linkAnalysis.External != 2 {
		t.Errorf("Expected 2 external links, got %d", linkAnalysis.External)
	}

	if linkAnalysis.Inaccessible != 0 {
		t.Errorf("Expected 0 inaccessible links, got %d", linkAnalysis.Inaccessible)
	}

	if len(linkAnalysis.BrokenLinks) != 0 {
		t.Errorf("Expected 0 broken links, got %d", len(linkAnalysis.BrokenLinks))
	}
}

func TestIsInternalLink(t *testing.T) {
	parser := &htmlParser{}
	baseURL := "https://example.com"

	tests := []struct {
		href     string
		expected bool
	}{
		{"/page", true},
		{"page", true},
		{"#fragment", true},
		{"?query=value", true},
		{"https://example.com/page", true},
		{"https://other.com/page", false},
		{"http://example.com/page", true},
		{"mailto:test@example.com", false},
		{"tel:+1234567890", false},
	}

	for _, test := range tests {
		result := parser.isInternalLink(test.href, baseURL)
		if result != test.expected {
			t.Errorf("isInternalLink(%q, %q) = %v, expected %v", test.href, baseURL, result, test.expected)
		}
	}
}

func TestCheckLinkAccessibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/accessible":
			w.WriteHeader(http.StatusOK)
		case "/redirect":
			w.WriteHeader(http.StatusMovedPermanently)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// create parser with proper HTTP client
	wrappedClient := NewHTTPClient(http.DefaultClient)
	parser := NewHTMLParser(wrappedClient)

	tests := []struct {
		href     string
		baseURL  string
		expected bool
	}{
		{"#fragment", server.URL, true},
		{"?query=value", server.URL, true},
		{"mailto:test@example.com", server.URL, true},
		{"tel:+1234567890", server.URL, true},
		{"/accessible", server.URL, true},
		{"/notfound", server.URL, false},
		{"accessible", server.URL, true},
		{"notfound", server.URL, false},
	}

	for _, test := range tests {
		result := parser.(*htmlParser).checkLinkAccessibility(test.href, test.baseURL)
		if result != test.expected {
			t.Errorf("checkLinkAccessibility(%q, %q) = %v, expected %v", test.href, test.baseURL, result, test.expected)
		}
	}
}
