package services

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"webpage-analyzer/internal/domain/entities"

	"golang.org/x/net/html"
)

type AnalyzerService interface {
	AnalyzeURL(ctx context.Context, targetURL string) (*entities.AnalysisResult, error)
	ValidateURL(url string) error
}

type analyzerService struct {
	httpClient HTTPClient
	parser     HTMLParser
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
	GetWithContext(ctx context.Context, url string) (*http.Response, error)
}

// httpClientWrapper wraps the standard http.Client to implement our HTTPClient interface
type httpClientWrapper struct {
	client *http.Client
}

func (w *httpClientWrapper) Get(url string) (*http.Response, error) {
	return w.client.Get(url)
}

func (w *httpClientWrapper) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return w.client.Do(req)
}

// NewHTTPClient creates a new HTTPClient wrapper
func NewHTTPClient(client *http.Client) HTTPClient {
	return &httpClientWrapper{client: client}
}

type HTMLParser interface {
	Parse(content string, baseURL string) (*ParsedHTML, error)
}

type ParsedHTML struct {
	HTMLVersion   string
	Title         string
	Headings      map[string]int
	Links         []Link
	HasLoginForm  bool
	ContentLength int64
}

type Link struct {
	URL          string
	IsInternal   bool
	IsAccessible bool
}

func NewAnalyzerService(httpClient HTTPClient, parser HTMLParser) AnalyzerService {
	return &analyzerService{
		httpClient: httpClient,
		parser:     parser,
	}
}

func (s *analyzerService) ValidateURL(targetURL string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must include scheme and host")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS schemes are supported")
	}

	return nil
}

func (s *analyzerService) createDetailedError(err error, targetURL string) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *url.Error:
		if e.Timeout() {
			return fmt.Errorf("connection timeout exceeded while accessing %s", targetURL)
		}
		if netErr, ok := e.Err.(net.Error); ok {
			if netErr.Timeout() {
				return fmt.Errorf("connection timeout exceeded while accessing %s", targetURL)
			}
		}
		if strings.Contains(e.Error(), "no such host") {
			return fmt.Errorf("domain not found: %s", targetURL)
		}
		if strings.Contains(e.Error(), "connection refused") {
			return fmt.Errorf("connection refused by server: %s", targetURL)
		}
		if strings.Contains(e.Error(), "network is unreachable") {
			return fmt.Errorf("network is unreachable: %s", targetURL)
		}
		return fmt.Errorf("network error while accessing %s: %v", targetURL, e.Err)
	case net.Error:
		if e.Timeout() {
			return fmt.Errorf("connection timeout exceeded while accessing %s", targetURL)
		}
		return fmt.Errorf("network error while accessing %s: %v", targetURL, e)
	default:
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return fmt.Errorf("context deadline exceeded while accessing %s", targetURL)
		}
		if strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("connection timeout exceeded while accessing %s", targetURL)
		}
		if strings.Contains(err.Error(), "refused") {
			return fmt.Errorf("connection refused by server: %s", targetURL)
		}
		if strings.Contains(err.Error(), "no such host") {
			return fmt.Errorf("domain not found: %s", targetURL)
		}
		return fmt.Errorf("failed to access %s: %v", targetURL, err)
	}
}

func (s *analyzerService) AnalyzeURL(ctx context.Context, targetURL string) (*entities.AnalysisResult, error) {
	startTime := time.Now()

	if err := s.ValidateURL(targetURL); err != nil {
		return nil, err
	}

	// Create context with deadline for the main request
	requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := s.httpClient.GetWithContext(requestCtx, targetURL)
	if err != nil {
		return nil, s.createDetailedError(err, targetURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorMsg string
		switch resp.StatusCode {
		case 403:
			errorMsg = "website blocked access (likely bot protection)"
		case 404:
			errorMsg = "page not found"
		case 429:
			errorMsg = "rate limit exceeded"
		case 500, 502, 503, 504:
			errorMsg = "server error"
		default:
			errorMsg = resp.Status
		}
		return &entities.AnalysisResult{
			StatusCode: resp.StatusCode,
			LoadTime:   time.Since(startTime),
		}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errorMsg)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	parsed, err := s.parser.Parse(string(content), targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	linkAnalysis := s.analyzeLinkAccessibility(ctx, parsed.Links)

	return &entities.AnalysisResult{
		HTMLVersion:   parsed.HTMLVersion,
		Title:         parsed.Title,
		Headings:      parsed.Headings,
		Links:         linkAnalysis,
		HasLoginForm:  parsed.HasLoginForm,
		LoadTime:      time.Since(startTime),
		ContentLength: parsed.ContentLength,
		StatusCode:    resp.StatusCode,
	}, nil
}

func (s *analyzerService) analyzeLinkAccessibility(ctx context.Context, links []Link) entities.LinkAnalysis {
	analysis := entities.LinkAnalysis{
		BrokenLinks:   make([]string, 0),
		ExternalHosts: make([]string, 0),
	}

	hostMap := make(map[string]bool)

	for _, link := range links {
		if link.IsInternal {
			analysis.Internal++
		} else {
			analysis.External++
			if linkU, err := url.Parse(link.URL); err == nil && !hostMap[linkU.Host] {
				analysis.ExternalHosts = append(analysis.ExternalHosts, linkU.Host)
				hostMap[linkU.Host] = true
			}
		}

		if !link.IsAccessible {
			analysis.Inaccessible++
			analysis.BrokenLinks = append(analysis.BrokenLinks, link.URL)
		}
	}

	return analysis
}

type htmlParser struct {
	httpClient HTTPClient
}

func NewHTMLParser(httpClient HTTPClient) HTMLParser {
	return &htmlParser{httpClient: httpClient}
}

func (p *htmlParser) Parse(content string, baseURL string) (*ParsedHTML, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	parsed := &ParsedHTML{
		Headings:      make(map[string]int),
		Links:         make([]Link, 0, 50),
		ContentLength: int64(len(content)),
	}

	parsed.HTMLVersion = p.extractHTMLVersion(doc)
	parsed.Title = p.extractTitle(doc)
	parsed.Headings = p.extractHeadings(doc)
	parsed.Links = p.extractLinks(doc, baseURL)
	parsed.HasLoginForm = p.hasLoginForm(doc)

	return parsed, nil
}

func (p *htmlParser) extractHTMLVersion(doc *html.Node) string {
	doctypeMap := map[string]string{
		"html 4.01 strict":        "HTML 4.01 Strict",
		"html 4.01//strict":       "HTML 4.01 Strict",
		"html 4.01 transitional":  "HTML 4.01 Transitional",
		"html 4.01//transitional": "HTML 4.01 Transitional",
		"html 4.01 frameset":      "HTML 4.01 Frameset",
		"html 4.01//frameset":     "HTML 4.01 Frameset",
		"html 4.0":                "HTML 4.0",
		"html 3.2":                "HTML 3.2",
		"html 2.0":                "HTML 2.0",
		"xhtml 1.1":               "XHTML 1.1",
		"xhtml 1.0 strict":        "XHTML 1.0 Strict",
		"xhtml 1.0//strict":       "XHTML 1.0 Strict",
		"xhtml 1.0 transitional":  "XHTML 1.0 Transitional",
		"xhtml 1.0//transitional": "XHTML 1.0 Transitional",
		"xhtml 1.0 frameset":      "XHTML 1.0 Frameset",
		"xhtml 1.0//frameset":     "XHTML 1.0 Frameset",
		"xhtml basic":             "XHTML Basic",
		"xhtml mobile":            "XHTML Mobile Profile",
		"xhtml":                   "XHTML",
		"html":                    "HTML",
	}

	var findDoctype func(*html.Node) string
	findDoctype = func(n *html.Node) string {
		if n.Type == html.DoctypeNode {
			doctype := strings.ToLower(n.Data)
			for pattern, version := range doctypeMap {
				if strings.Contains(doctype, pattern) {
					return version
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := findDoctype(c); result != "" {
				return result
			}
		}
		return ""
	}

	version := findDoctype(doc)
	if version == "" {
		if p.hasHTML5Features(doc) {
			return "HTML5"
		}
		return "Unknown/No DOCTYPE"
	}
	return version
}

func (p *htmlParser) extractTitle(doc *html.Node) string {
	var findTitle func(*html.Node) string
	findTitle = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				return strings.TrimSpace(n.FirstChild.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if title := findTitle(c); title != "" {
				return title
			}
		}
		return ""
	}
	return findTitle(doc)
}

func (p *htmlParser) extractHeadings(doc *html.Node) map[string]int {
	headings := make(map[string]int)
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				headings[n.Data]++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return headings
}

func (p *htmlParser) extractLinks(doc *html.Node, baseURL string) []Link {
	links := make([]Link, 0, 100)
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" {
					link := Link{
						URL:          attr.Val,
						IsInternal:   p.isInternalLink(attr.Val, baseURL),
						IsAccessible: p.checkLinkAccessibility(attr.Val, baseURL),
					}
					links = append(links, link)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return links
}

func (p *htmlParser) isInternalLink(href string, baseURL string) bool {
	// Fragment links and query parameters are always internal
	if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "?") {
		return true
	}

	// Relative paths starting with / are internal
	if strings.HasPrefix(href, "/") {
		return true
	}

	// Parse the href
	hrefURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	// If no scheme or host, it's a relative path (internal)
	if hrefURL.Scheme == "" && hrefURL.Host == "" {
		return true
	}

	// Parse the base URL to compare hosts
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	// Compare hosts - if they match, it's internal
	return hrefURL.Host == baseURLParsed.Host
}

func (p *htmlParser) checkLinkAccessibility(href string, baseURL string) bool {
	// Fragment links, query parameters, mailto, and tel links are always considered accessible
	if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "?") ||
		strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
		return true
	}

	// Parse the href
	hrefURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	// If no scheme or host, it's a relative path - resolve it against base URL
	if hrefURL.Scheme == "" && hrefURL.Host == "" {
		baseURLParsed, err := url.Parse(baseURL)
		if err != nil {
			return false
		}
		resolvedURL := baseURLParsed.ResolveReference(hrefURL)
		return p.checkHTTPLink(resolvedURL.String())
	}

	// If it's a relative path starting with /
	if strings.HasPrefix(href, "/") {
		baseURLParsed, err := url.Parse(baseURL)
		if err != nil {
			return false
		}
		// Create a new URL with the base scheme and host, but the href path
		resolvedURL := &url.URL{
			Scheme: baseURLParsed.Scheme,
			Host:   baseURLParsed.Host,
			Path:   href,
		}
		return p.checkHTTPLink(resolvedURL.String())
	}

	// Only check HTTP/HTTPS links
	if hrefURL.Scheme != "http" && hrefURL.Scheme != "https" {
		return true // Consider non-HTTP schemes as accessible
	}

	return p.checkHTTPLink(href)
}

// checkHTTPLink performs the actual HTTP check with context deadline
func (p *htmlParser) checkHTTPLink(url string) bool {
	// Create context with deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	// Create HTTP client
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as accessible
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (p *htmlParser) hasLoginForm(doc *html.Node) bool {
	loginKeywords := map[string]bool{
		"login": true, "signin": true, "sign-in": true, "sign_in": true, "log-in": true, "log_in": true,
		"sign in": true, "log in": true, "logon": true, "log on": true,
		"password": true, "passwd": true, "pwd": true, "pass": true,
		"username": true, "userid": true, "user_id": true,
		"authenticate": true, "authentication": true,
		"credentials": true, "credential": true,
		"forgot password": true, "reset password": true, "password reset": true,
		"remember me": true, "stay logged in": true,
	}

	loginInputNames := map[string]bool{
		"username": true, "userid": true, "user_id": true,
		"password": true, "passwd": true, "pwd": true, "pass": true, "passphrase": true,
		"login": true, "signin": true, "authenticate": true,
		"remember": true, "remember_me": true, "stay_logged_in": true,
	}

	var hasPasswordField bool
	var hasLoginContext bool

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "input" {
				var inputType, inputName, inputId, inputClass string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "type":
						inputType = strings.ToLower(attr.Val)
					case "name":
						inputName = strings.ToLower(attr.Val)
					case "id":
						inputId = strings.ToLower(attr.Val)
					case "class":
						inputClass = strings.ToLower(attr.Val)
					}
				}

				if inputType == "password" {
					hasPasswordField = true
				}
				for loginName := range loginInputNames {
					if strings.Contains(inputName, loginName) || strings.Contains(inputId, loginName) || strings.Contains(inputClass, loginName) {
						hasLoginContext = true
					}
				}
			}

			if n.Data == "form" {
				for _, attr := range n.Attr {
					if attr.Key == "action" || attr.Key == "id" || attr.Key == "class" || attr.Key == "name" {
						value := strings.ToLower(attr.Val)
						for keyword := range loginKeywords {
							if strings.Contains(value, keyword) {
								hasLoginContext = true
							}
						}
					}
				}
			}

			if n.Data == "button" || n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "id" || attr.Key == "class" || attr.Key == "name" || attr.Key == "value" {
						value := strings.ToLower(attr.Val)
						for keyword := range loginKeywords {
							if strings.Contains(value, keyword) {
								hasLoginContext = true
							}
						}
					}
				}
			}

			if n.Data == "h1" || n.Data == "h2" || n.Data == "h3" || n.Data == "h4" || n.Data == "h5" || n.Data == "h6" ||
				n.Data == "label" || n.Data == "span" || n.Data == "div" || n.Data == "p" || n.Data == "a" ||
				n.Data == "button" || n.Data == "title" || n.Data == "legend" {
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					text := strings.ToLower(strings.TrimSpace(n.FirstChild.Data))
					for keyword := range loginKeywords {
						if strings.Contains(text, keyword) {
							hasLoginContext = true
						}
					}
				}
			}

		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	return hasPasswordField && hasLoginContext
}

func (p *htmlParser) hasHTML5Features(doc *html.Node) bool {
	html5Elements := map[string]bool{
		"article": true, "aside": true, "audio": true, "canvas": true,
		"datalist": true, "details": true, "embed": true, "figcaption": true,
		"figure": true, "footer": true, "header": true, "hgroup": true,
		"keygen": true, "mark": true, "meter": true, "nav": true,
		"output": true, "progress": true, "rp": true, "rt": true,
		"ruby": true, "section": true, "source": true, "summary": true,
		"time": true, "track": true, "video": true, "wbr": true,
	}

	html5InputTypes := map[string]bool{
		"email": true, "url": true, "tel": true, "search": true,
		"number": true, "range": true, "date": true, "time": true,
		"datetime": true, "datetime-local": true, "month": true,
		"week": true, "color": true,
	}

	html5Attributes := map[string]bool{
		"contenteditable": true, "draggable": true, "hidden": true, "spellcheck": true,
	}

	var traverse func(*html.Node) bool
	traverse = func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			if html5Elements[n.Data] {
				return true
			}

			if n.Data == "input" {
				for _, attr := range n.Attr {
					if attr.Key == "type" && html5InputTypes[attr.Val] {
						return true
					}
				}
			}

			for _, attr := range n.Attr {
				if html5Attributes[attr.Key] {
					return true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if traverse(c) {
				return true
			}
		}
		return false
	}

	return traverse(doc)
}
