package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	semaphore  chan struct{} // limit concurrent checks
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
	GetWithContext(ctx context.Context, url string) (*http.Response, error)
}

type httpClientWrapper struct {
	client *http.Client
}

func (w *httpClientWrapper) Get(url string) (*http.Response, error) {
	return w.client.Get(url)
}

func (w *httpClientWrapper) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, HTTPMethodGET, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func NewHTTPClient(client *http.Client) HTTPClient {
	return &httpClientWrapper{client: client}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type HTMLParser interface {
	Parse(content string, baseURL string) (*ParsedHTML, error)
}

type ParsedHTML struct {
	HTMLVersion   string         `json:"html_version"`
	Title         string         `json:"title"`
	Headings      map[string]int `json:"headings"`
	Links         []Link         `json:"links"`
	HasLoginForm  bool           `json:"has_login_form"`
	ContentLength int64          `json:"content_length"`
}

type Link struct {
	URL          string `json:"url"`
	IsInternal   bool   `json:"is_internal"`
	IsAccessible bool   `json:"is_accessible"`
}

func NewAnalyzerService(httpClient HTTPClient, parser HTMLParser, maxConcurrentChecks int) AnalyzerService {
	if maxConcurrentChecks <= 0 {
		maxConcurrentChecks = 10
	}

	return &analyzerService{
		httpClient: httpClient,
		parser:     parser,
		semaphore:  make(chan struct{}, maxConcurrentChecks),
	}
}

func (s *analyzerService) ValidateURL(targetURL string) error {
	if targetURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if len(targetURL) > MaxURLLength {
		return fmt.Errorf("URL too long (max %d characters)", MaxURLLength)
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must include scheme and host")
	}

	if !contains(SupportedSchemes, u.Scheme) {
		return fmt.Errorf("only %v schemes are supported", SupportedSchemes)
	}

	// allow localhost for testing
	_ = u.Hostname()

	if strings.Contains(u.Hostname(), "..") {
		return fmt.Errorf("invalid hostname format")
	}

	return nil
}

func (s *analyzerService) createDetailedError(err error, targetURL string) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "context deadline exceeded") {
		return fmt.Errorf("context deadline exceeded while accessing %s", targetURL)
	}

	if strings.Contains(err.Error(), "context canceled") {
		return fmt.Errorf("request was canceled while accessing %s", targetURL)
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
		return s.classifyNetworkError(e.Error(), targetURL)

	case net.Error:
		if e.Timeout() {
			return fmt.Errorf("connection timeout exceeded while accessing %s", targetURL)
		}
		return fmt.Errorf("network error while accessing %s: %v", targetURL, e)

	default:
		return s.classifyNetworkError(err.Error(), targetURL)
	}
}

func (s *analyzerService) classifyNetworkError(errorMsg, targetURL string) error {
	errorMsg = strings.ToLower(errorMsg)

	switch {
	case strings.Contains(errorMsg, "no such host") || strings.Contains(errorMsg, "name resolution"):
		return fmt.Errorf("domain not found: %s", targetURL)
	case strings.Contains(errorMsg, "connection refused"):
		return fmt.Errorf("connection refused by server: %s", targetURL)
	case strings.Contains(errorMsg, "network is unreachable"):
		return fmt.Errorf("network is unreachable: %s", targetURL)
	case strings.Contains(errorMsg, "timeout"):
		return fmt.Errorf("connection timeout exceeded while accessing %s", targetURL)
	case strings.Contains(errorMsg, "tls") || strings.Contains(errorMsg, "certificate"):
		return fmt.Errorf("SSL/TLS error while accessing %s", targetURL)
	default:
		return fmt.Errorf("network error while accessing %s: %s", targetURL, errorMsg)
	}
}

func (s *analyzerService) AnalyzeURL(ctx context.Context, targetURL string) (*entities.AnalysisResult, error) {
	startTime := time.Now()

	if err := s.ValidateURL(targetURL); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	// change timeout here if needed
	requestCtx, cancel := context.WithTimeout(ctx, DefaultRequestTimeout)
	defer cancel()

	resp, err := s.httpClient.GetWithContext(requestCtx, targetURL)
	if err != nil {
		return nil, s.createDetailedError(err, targetURL)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log close error but don't fail the analysis
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		errorMsg := s.getHTTPStatusMessage(resp.StatusCode)
		return &entities.AnalysisResult{
			StatusCode: resp.StatusCode,
			LoadTime:   time.Since(startTime),
		}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errorMsg)
	}

	content, err := io.ReadAll(io.LimitReader(resp.Body, MaxContentSize))
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

func (s *analyzerService) getHTTPStatusMessage(statusCode int) string {
	switch statusCode {
	case 400:
		return "bad request"
	case 401:
		return "authentication required"
	case 403:
		return "website blocked access (likely bot protection)"
	case 404:
		return "page not found"
	case 429:
		return "rate limit exceeded"
	case 500, 502, 503, 504:
		return "server error"
	default:
		return fmt.Sprintf("HTTP %d", statusCode)
	}
}

func (s *analyzerService) analyzeLinkAccessibility(ctx context.Context, links []Link) entities.LinkAnalysis {
	analysis := entities.LinkAnalysis{
		BrokenLinks:   make([]string, 0),
		ExternalHosts: make([]string, 0),
	}

	hostMap := make(map[string]bool)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		go func(l Link) {
			defer wg.Done()

			// limit concurrent checks
			select {
			case s.semaphore <- struct{}{}:
				defer func() { <-s.semaphore }()
			case <-ctx.Done():
				return
			}

			mu.Lock()
			if l.IsInternal {
				analysis.Internal++
			} else {
				analysis.External++
				if linkU, err := url.Parse(l.URL); err == nil && !hostMap[linkU.Host] {
					analysis.ExternalHosts = append(analysis.ExternalHosts, linkU.Host)
					hostMap[linkU.Host] = true
				}
			}

			if !l.IsAccessible {
				analysis.Inaccessible++
				analysis.BrokenLinks = append(analysis.BrokenLinks, l.URL)
			}
			mu.Unlock()
		}(link)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	return analysis
}

type htmlParser struct {
	httpClient HTTPClient
	urlCache   map[string]bool // cache results
	mu         sync.RWMutex
}

func NewHTMLParser(httpClient HTTPClient) HTMLParser {
	return &htmlParser{
		httpClient: httpClient,
		urlCache:   make(map[string]bool),
	}
}

func (p *htmlParser) Parse(content string, baseURL string) (*ParsedHTML, error) {
	if content == "" {
		return nil, fmt.Errorf("HTML content cannot be empty")
	}

	if len(content) > MaxContentSize {
		return nil, fmt.Errorf("HTML content too large (max %d bytes)", MaxContentSize)
	}

	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
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
		if n.Type == html.ElementNode && n.Data == HTMLElementTitle {
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
	var traverse func(*html.Node, int)
	traverse = func(n *html.Node, depth int) {
		if depth > MaxHTMLDepth {
			return
		}
		if n.Type == html.ElementNode {
			switch n.Data {
			case HTMLElementH1, HTMLElementH2, HTMLElementH3, HTMLElementH4, HTMLElementH5, HTMLElementH6:
				headings[n.Data]++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c, depth+1)
		}
	}
	traverse(doc, 0)
	return headings
}

func (p *htmlParser) extractLinks(doc *html.Node, baseURL string) []Link {
	links := make([]Link, 0, 100)
	var traverse func(*html.Node, int)

	traverse = func(n *html.Node, depth int) {
		if depth > MaxHTMLDepth {
			return
		}
		if n.Type == html.ElementNode && n.Data == HTMLElementA {
			for _, attr := range n.Attr {
				if attr.Key == HTMLAttrHref && attr.Val != "" {
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
			traverse(c, depth+1)
		}
	}
	traverse(doc, 0)
	return links
}

func (p *htmlParser) isInternalLink(href string, baseURL string) bool {

	if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "?") {
		return true
	}

	if strings.HasPrefix(href, "/") {
		return true
	}

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
	if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "?") ||
		strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
		return true
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	var fullURL string

	if hrefURL.Scheme == "" && hrefURL.Host == "" {
		baseURLParsed, err := url.Parse(baseURL)
		if err != nil {
			return false
		}
		resolvedURL := baseURLParsed.ResolveReference(hrefURL)
		fullURL = resolvedURL.String()
	} else if strings.HasPrefix(href, "/") {
		baseURLParsed, err := url.Parse(baseURL)
		if err != nil {
			return false
		}
		resolvedURL := &url.URL{
			Scheme: baseURLParsed.Scheme,
			Host:   baseURLParsed.Host,
			Path:   href,
		}
		fullURL = resolvedURL.String()
	} else {
		if !contains(SupportedSchemes, hrefURL.Scheme) {
			return true
		}
		fullURL = href
	}

	// check cache first
	p.mu.RLock()
	if accessible, exists := p.urlCache[fullURL]; exists {
		p.mu.RUnlock()
		return accessible
	}
	p.mu.RUnlock()

	accessible := p.checkHTTPLink(fullURL)

	// cache result
	p.mu.Lock()
	p.urlCache[fullURL] = accessible
	p.mu.Unlock()

	return accessible
}

func (p *htmlParser) checkHTTPLink(url string) bool {
	// change timeout here if needed
	ctx, cancel := context.WithTimeout(context.Background(), DefaultLinkCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, HTTPMethodHEAD, url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := p.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return false
	}

	if resp == nil {
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log close error but don't fail the analysis
			// This is best-effort cleanup - network issues shouldn't break analysis
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

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

	var traverse func(*html.Node, int)
	traverse = func(n *html.Node, depth int) {
		if depth > MaxHTMLDepth {
			return
		}
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

			if n.Data == HTMLElementButton || n.Data == HTMLElementA {
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

			if contains(AccessibilityElements, n.Data) {
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
			traverse(c, depth+1)
		}
	}

	traverse(doc, 0)

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
		LinkTypeEmail: true, LinkTypeURL: true, LinkTypeTel: true, LinkTypeSearch: true,
		"number": true, "range": true, "date": true, "time": true,
		"datetime": true, "datetime-local": true, "month": true,
		"week": true, "color": true,
	}

	html5Attributes := map[string]bool{
		"contenteditable": true, "draggable": true, "hidden": true, "spellcheck": true,
	}

	var traverse func(*html.Node, int) bool
	traverse = func(n *html.Node, depth int) bool {
		if depth > MaxHTMLDepth {
			return false
		}
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
			if traverse(c, depth+1) {
				return true
			}
		}
		return false
	}

	return traverse(doc, 0)
}
