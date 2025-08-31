package services

import (
	"context"
	"fmt"
	"io"
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
}

type HTMLParser interface {
	Parse(content string) (*ParsedHTML, error)
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

func (s *analyzerService) AnalyzeURL(ctx context.Context, targetURL string) (*entities.AnalysisResult, error) {
	startTime := time.Now()

	if err := s.ValidateURL(targetURL); err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &entities.AnalysisResult{
			StatusCode: resp.StatusCode,
			LoadTime:   time.Since(startTime),
		}, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	parsed, err := s.parser.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	linkAnalysis := s.analyzeLinkAccessibility(ctx, parsed.Links, targetURL)

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

func (s *analyzerService) analyzeLinkAccessibility(ctx context.Context, links []Link, baseURL string) entities.LinkAnalysis {
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
			linkU, _ := url.Parse(link.URL)
			if linkU != nil && !hostMap[linkU.Host] {
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

func (p *htmlParser) Parse(content string) (*ParsedHTML, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	parsed := &ParsedHTML{
		Headings:      make(map[string]int),
		Links:         make([]Link, 0),
		ContentLength: int64(len(content)),
	}

	parsed.HTMLVersion = p.extractHTMLVersion(doc)
	parsed.Title = p.extractTitle(doc)
	parsed.Headings = p.extractHeadings(doc)
	parsed.Links = p.extractLinks(doc)
	parsed.HasLoginForm = p.hasLoginForm(doc)

	return parsed, nil
}

func (p *htmlParser) extractHTMLVersion(doc *html.Node) string {
	var findDoctype func(*html.Node) string
	findDoctype = func(n *html.Node) string {
		if n.Type == html.DoctypeNode {
			doctype := strings.ToLower(n.Data)
			if strings.Contains(doctype, "html") {
				if strings.Contains(doctype, "4.01") {
					return "HTML 4.01"
				} else if strings.Contains(doctype, "xhtml") {
					return "XHTML"
				}
			}
			return "HTML5"
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
		return "HTML5"
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

func (p *htmlParser) extractLinks(doc *html.Node) []Link {
	links := make([]Link, 0)
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" {
					link := Link{
						URL:          attr.Val,
						IsInternal:   p.isInternalLink(attr.Val),
						IsAccessible: true,
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

func (p *htmlParser) isInternalLink(href string) bool {
	if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "?") {
		return true
	}

	u, err := url.Parse(href)
	if err != nil {
		return false
	}

	return u.Host == ""
}

func (p *htmlParser) hasLoginForm(doc *html.Node) bool {
	var traverse func(*html.Node) bool
	traverse = func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			if n.Data == "input" {
				for _, attr := range n.Attr {
					if attr.Key == "type" && attr.Val == "password" {
						return true
					}
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
