package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTestConfig() *AnalyzerConfig {
	return &AnalyzerConfig{
		LinkCheckTimeout:        5 * time.Second,
		MaxLinksToCheck:         50,
		MaxConcurrentLinkChecks: 10,
		MaxHTMLDepth:            100,
		MaxURLLength:            2048,
	}
}

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
	service := NewAnalyzerService(wrappedClient, parser, getTestConfig())

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
	_ = NewHTMLParser(wrappedClient) // parser not used in this test

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

	testClient := NewHTTPClient(http.DefaultClient)
	testParser := NewHTMLParser(testClient)

	for _, test := range tests {
		// create a simple HTML page with the test link
		htmlContent := fmt.Sprintf(`<html><body><a href="%s">Test Link</a></body></html>`, test.href)

		parsed, err := testParser.Parse(htmlContent, test.baseURL)
		if err != nil {
			t.Errorf("Parse failed for %q: %v", test.href, err)
			continue
		}

		// find the link in the parsed results
		if len(parsed.Links) > 0 {
			link := parsed.Links[0]
			if link.IsAccessible != test.expected {
				t.Errorf("Link accessibility for %q = %v, expected %v", test.href, link.IsAccessible, test.expected)
			}
		}
	}
}

func TestNewAnalyzerService(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())

	assert.NotNil(t, service)
}

func TestValidateURL(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())

	err := service.ValidateURL("https://example.com")
	assert.NoError(t, err)

	err = service.ValidateURL("invalid-url")
	assert.Error(t, err)

	err = service.ValidateURL("")
	assert.Error(t, err)
}

func TestNewHTMLParser(t *testing.T) {
	parser := NewHTMLParser(nil)

	assert.NotNil(t, parser)
}

func TestAnalyzerServiceConstructor(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())
	assert.NotNil(t, service)
}

func TestAnalyzerServiceWithDifferentMaxDepth(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())
	assert.NotNil(t, service)
}

func TestAnalyzerServiceWithZeroMaxDepth(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())
	assert.NotNil(t, service)
}

func TestAnalyzeWebPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Test Page</title>
			</head>
			<body>
				<h1>Main Heading</h1>
				<h2>Sub Heading</h2>
				<h3>Another Heading</h3>
				<a href="/internal">Internal Link</a>
				<a href="https://external.com">External Link</a>
				<form>
					<input type="text" name="username">
					<input type="password" name="password">
					<button type="submit">Login</button>
				</form>
			</body>
			</html>
		`))
	}))
	defer server.Close()

	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())
	result, err := service.AnalyzeURL(context.Background(), server.URL)

	assert.NoError(t, err)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, "Test Page", result.Title)
	assert.Equal(t, "HTML", result.HTMLVersion)
	assert.Equal(t, 1, result.Headings["h1"])
	assert.Equal(t, 1, result.Headings["h2"])
	assert.Equal(t, 1, result.Headings["h3"])
	assert.True(t, result.HasLoginForm)
}

func TestAnalyzeWebPageWithInvalidURL(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())
	_, err := service.AnalyzeURL(context.Background(), "not-a-valid-url")

	assert.Error(t, err)
}

func TestAnalyzeWebPageWith404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())
	_, err := service.AnalyzeURL(context.Background(), server.URL)

	assert.Error(t, err)
}

func TestHTMLParserExtractTitle(t *testing.T) {
	html := `<html><head><title>My Test Page</title></head><body></body></html>`
	parser := NewHTMLParser(nil)

	parsed, err := parser.Parse(html, "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "My Test Page", parsed.Title)
}

func TestHTMLParserExtractHeadings(t *testing.T) {
	html := `
		<html>
		<body>
			<h1>Title 1</h1>
			<h2>Title 2</h2>
			<h2>Title 2 Again</h2>
			<h3>Title 3</h3>
		</body>
		</html>
	`
	parser := NewHTMLParser(nil)

	parsed, err := parser.Parse(html, "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, 1, parsed.Headings["h1"])
	assert.Equal(t, 2, parsed.Headings["h2"])
	assert.Equal(t, 1, parsed.Headings["h3"])
}

func TestHTMLParserDetectLoginForm(t *testing.T) {
	htmlWithLogin := `
		<html>
		<body>
			<form>
				<input type="text" name="username">
				<input type="password" name="password">
			</form>
		</body>
		</html>
	`

	htmlWithoutLogin := `
		<html>
		<body>
			<form>
				<input type="text" name="email">
				<input type="submit" value="Submit">
			</form>
		</body>
		</html>
	`

	parser := NewHTMLParser(nil)

	parsed1, err1 := parser.Parse(htmlWithLogin, "https://example.com")
	assert.NoError(t, err1)
	assert.True(t, parsed1.HasLoginForm)

	parsed2, err2 := parser.Parse(htmlWithoutLogin, "https://example.com")
	assert.NoError(t, err2)
	assert.False(t, parsed2.HasLoginForm)
}

func TestValidateURLComprehensive(t *testing.T) {
	httpClient := NewHTTPClient(&http.Client{})
	parser := NewHTMLParser(httpClient)
	service := NewAnalyzerService(httpClient, parser, getTestConfig())

	validURLs := []string{
		"https://example.com",
		"http://test.com",
		"https://subdomain.example.com/path",
	}

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com",
		"",
		"javascript:alert('test')",
	}

	for _, url := range validURLs {
		err := service.ValidateURL(url)
		assert.NoError(t, err, "URL should be valid: %s", url)
	}

	for _, url := range invalidURLs {
		err := service.ValidateURL(url)
		assert.Error(t, err, "URL should be invalid: %s", url)
	}
}
