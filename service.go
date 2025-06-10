package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func traverse(n *html.Node, domResults *DOMAnalysis, url *url.URL) (map[int]int, Links) {
	headingCounts := make(map[int]int)
	links := Links{
		Internal:     0,
		External:     0,
		Inaccessible: 0,
	}

	// Check current node
	if n.Type == html.ElementNode {
		// Check for headings
		if level, yes := isHeadingTag(n); yes {
			headingCounts[level]++
		}
		// Check for title
		if title, yes := isTitleTag(n); yes {
			domResults.Title = title
		}
		// Check for links
		if yes, internal, working := isLinkTag(n, url); yes {
			if internal {
				links.Internal++
			} else {
				links.External++
			}
			if !working {
				links.Inaccessible++
			}
		}
		// Check for login form
		if strings.ToLower(n.Data) == "form" {
			if isLoginForm(n) {
				domResults.LoginForm = true
			}
		}
	}

	// Process all child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childCounts, childLinks := traverse(c, domResults, url)
		// Merge child counts into parent counts
		for level, count := range childCounts {
			headingCounts[level] += count
		}
		// Merge child links into parent links
		links.Internal += childLinks.Internal
		links.External += childLinks.External
		links.Inaccessible += childLinks.Inaccessible
	}

	return headingCounts, links
}

func analyseDOM(reader io.Reader, baseUrl *url.URL) (*DOMAnalysis, error) {
	domResults := &DOMAnalysis{
		LoginForm: false, // Initialize explicitly
	}
	n, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	headingCounts, links := traverse(n, domResults, baseUrl)

	// Create final heading list with counts
	for level := 1; level <= 6; level++ {
		if count, ok := headingCounts[level]; ok {
			domResults.Headings = append(domResults.Headings, Heading{
				Tag:   fmt.Sprintf("h%d", level),
				Level: level,
				Count: count,
			})
		}
	}
	domResults.Links = links
	return domResults, nil
}

func extractHTMLVersion(r io.Reader) string {
	tokenizer := html.NewTokenizer(r)

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.DoctypeToken:
			tok := tokenizer.Token()
			doctype := tok.Data
			switch {
			case doctype == "html":
				return "HTML 5"
			case regexp.MustCompile(`(?i)html 4\.01`).MatchString(doctype):
				return "HTML 4.01"
			case regexp.MustCompile(`(?i)xhtml`).MatchString(doctype):
				return "XHTML"
			default:
				return "Unknown Doctype: " + doctype
			}
		case html.ErrorToken:
			return "No Doctype Found"
		}
	}
}

func isHeadingTag(n *html.Node) (int, bool) {
	var tag = strings.ToLower(n.Data)
	if strings.HasPrefix(tag, "h") && len(tag) == 2 {
		if level, err := strconv.Atoi(string(tag[1])); err == nil && level >= 1 && level <= 6 {
			return level, true
		}
	}
	return 0, false
}

func isTitleTag(n *html.Node) (string, bool) {
	if strings.ToLower(n.Data) == "title" && n.FirstChild != nil {
		if n.FirstChild.Type == html.TextNode {
			return n.FirstChild.Data, true
		}
	}
	return "", false
}

func isLinkTag(n *html.Node, base *url.URL) (isLink bool, isInternal, isWorking bool) {
	if strings.ToLower(n.Data) != "a" {
		return false, false, false
	}
	var link string
	for _, attr := range n.Attr {
		if strings.ToLower(attr.Key) == "href" {
			link = strings.TrimSpace(attr.Val)

			if link == "" || strings.HasPrefix(link, "#") {
				return true, true, true // internal, but skip checking
			}

			// Parse and resolve the link
			linkURL, err := url.Parse(link)
			if err != nil {
				return true, false, false // Invalid URL
			}
			resolvedURL := base.ResolveReference(linkURL)

			isInternal = resolvedURL.Host == base.Host || resolvedURL.Host == ""

			// Check if the link is working
			resp, err := http.Head(resolvedURL.String())
			if err != nil || resp.StatusCode >= 400 {
				// Try GET if HEAD fails
				resp, err = http.Get(resolvedURL.String())
				if err != nil || resp.StatusCode >= 400 {
					return true, isInternal, false
				}
			}
			defer resp.Body.Close()

			return true, isInternal, true
		}
	}
	return false, false, false // No href found, not a valid link
}

func isLoginForm(n *html.Node) bool {
	if strings.ToLower(n.Data) != "form" {
		return false
	}

	hasPassword := false
	hasEmailOrUsername := false
	hasSubmit := false

	var traverseForm func(*html.Node)
	traverseForm = func(node *html.Node) {
		if node.Type == html.ElementNode {
			// Check for password input
			if strings.ToLower(node.Data) == "input" {
				for _, attr := range node.Attr {
					attrKey := strings.ToLower(attr.Key)
					attrVal := strings.ToLower(attr.Val)

					if attrKey == "type" {
						if attrVal == "password" {
							hasPassword = true
						} else if attrVal == "email" || attrVal == "text" {
							// Check if this input is likely for username/email
							for _, a := range node.Attr {
								if strings.ToLower(a.Key) == "name" || strings.ToLower(a.Key) == "id" || strings.ToLower(a.Key) == "placeholder" {
									val := strings.ToLower(a.Val)
									if strings.Contains(val, "email") ||
										strings.Contains(val, "username") ||
										strings.Contains(val, "login") ||
										strings.Contains(val, "pass") ||
										strings.Contains(val, "phone") ||
										strings.Contains(val, "mobile") {
										hasEmailOrUsername = true
									}
								}
							}
						} else if attrVal == "submit" || attrVal == "button" {
							hasSubmit = true
						}
					}
				}
			}
			// Also check for button elements
			if strings.ToLower(node.Data) == "button" {
				hasSubmit = true
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverseForm(c)
		}
	}

	traverseForm(n)
	return hasPassword && (hasEmailOrUsername || hasSubmit)
}

// Validate URL
func ValidateURL(urlStr string) bool {
	re := regexp.MustCompile(`^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`)
	return re.MatchString(urlStr)
}

// Read page content
func ReadPageContent(urlStr string) (*PageAnalysis, error) {
	result := &PageAnalysis{}

	// Add https:// if not present
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		// Check for specific network errors
		if strings.Contains(err.Error(), "no such host") {
			return nil, fmt.Errorf("domain not found: %s", urlStr)
		}
		if strings.Contains(err.Error(), "connection refused") {
			return nil, fmt.Errorf("connection refused: server is not responding")
		}
		if strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("connection timeout: server took too long to respond")
		}
		// For any other unknown errors
		return nil, fmt.Errorf("unknown error while fetching URL: %v", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode >= 400 {
		statusText := http.StatusText(resp.StatusCode)
		if statusText == "" {
			statusText = "Unknown Status"
		}
		return nil, fmt.Errorf("HTTP error: received status code %d (%s)", resp.StatusCode, statusText)
	}

	// Read response body into memory
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Detect HTML version from one reader
	result.HTMLVersion = extractHTMLVersion(bytes.NewReader(bodyBytes))
	// Read DOM from another reader
	parsedUrl, _ := url.Parse(urlStr)
	da, err := analyseDOM(bytes.NewReader(bodyBytes), parsedUrl)
	if err != nil {
		return nil, fmt.Errorf("error analysing DOM: %w", err)
	}
	result.DOMAnalysis = *da
	return result, nil
}
