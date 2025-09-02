package services

import "time"

const (
	MaxContentSize             = 10 * 1024 * 1024 // 10MB
	MaxURLLength               = 2048
	MaxHTMLDepth               = 100
	DefaultMaxConcurrentChecks = 10
	DefaultRequestTimeout      = 30 * time.Second
	DefaultLinkCheckTimeout    = 10 * time.Second
	UserAgent                  = "WebPageAnalyzer/1.0"

	// HTTP methods
	HTTPMethodGET  = "GET"
	HTTPMethodHEAD = "HEAD"

	// HTML elements
	HTMLElementTitle  = "title"
	HTMLElementA      = "a"
	HTMLElementH1     = "h1"
	HTMLElementH2     = "h2"
	HTMLElementH3     = "h3"
	HTMLElementH4     = "h4"
	HTMLElementH5     = "h5"
	HTMLElementH6     = "h6"
	HTMLElementButton = "button"
	HTMLElementLabel  = "label"
	HTMLElementSpan   = "span"
	HTMLElementDiv    = "div"
	HTMLElementP      = "p"
	HTMLElementLegend = "legend"

	// HTML attributes
	HTMLAttrHref = "href"

	// Link types
	LinkTypeEmail  = "email"
	LinkTypeURL    = "url"
	LinkTypeTel    = "tel"
	LinkTypeSearch = "search"
)

var (
	SupportedSchemes = []string{"http", "https"}

	HeadingElements = []string{HTMLElementH1, HTMLElementH2, HTMLElementH3, HTMLElementH4, HTMLElementH5, HTMLElementH6}

	AccessibilityElements = []string{
		HTMLElementButton, HTMLElementA, HTMLElementH1, HTMLElementH2, HTMLElementH3,
		HTMLElementH4, HTMLElementH5, HTMLElementH6, HTMLElementLabel, HTMLElementSpan,
		HTMLElementDiv, HTMLElementP, HTMLElementLegend, HTMLElementTitle,
	}
)
