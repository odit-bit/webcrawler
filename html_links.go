package webcrawler

import (
	"context"
	"net/url"
	"regexp"

	"github.com/odit-bit/webcrawler/x/xpipe"
)

// link extractor only extract link that has the same host from the domain in html page

// var exclude_Extensions = []string{".jpg", ".jpeg", ".png", ".gif", ".ico", ".css", ".js", ".pdf"}

var (
	baseHrefRegex = regexp.MustCompile(`(?i)<base.*?href\s*?=\s*?"(.*?)\s*?"`)
	findLinkRegex = regexp.MustCompile(`(?i)<a.*?href\s*?=\s*?"\s*?(.*?)\s*?".*?>`)
	// nofollowRegex = regexp.MustCompile(`(?i)rel\s*?=\s*?"?nofollow"?`)
)

// TODO :  prevent double index (same destination)
// "example.com" and "example.com/" has same destination

// Process implements pipeline.Processor.
func ExtractURLs() xpipe.ProcessorFunc[*Resource] {
	return func(ctx context.Context, src *Resource) (*Resource, error) {
		payload := src
		relTo, err := url.Parse(trailingSlash(payload.URL))
		if err != nil {
			return nil, err
		}

		//search <base> tag n resolve to absolute url
		// <base href="XXX">
		content := payload.rawBuffer.String()
		baseMatch := baseHrefRegex.FindStringSubmatch(content)
		if len(baseMatch) == 2 {
			base := resolveURL(relTo, trailingSlash(baseMatch[1]))
			if base != nil {
				relTo = base
			}
		}

		//find unique set of link
		seenMap := make(map[string]struct{})
		// insert the payload.URL
		seenMap[relTo.String()] = struct{}{}

		for _, match := range findLinkRegex.FindAllStringSubmatch(content, -1) {
			link := resolveURL(relTo, match[1])

			if link == nil {
				continue
			}

			// Skip links to the other host
			if link.Host != relTo.Host {
				continue
			}

			// Truncate anchors and drop duplicates
			link.Fragment = ""
			linkStr := link.String()
			if _, seen := seenMap[linkStr]; seen {
				continue
			}

			// Skip URLs that point to files that cannot contain html content.
			if exclusionRegex.MatchString(linkStr) {
				continue
			}

			seenMap[linkStr] = struct{}{}
			payload.FoundURLs = append(payload.FoundURLs, linkStr)
		}

		return payload, nil

	}
}

func retainLink(srcHost string, link *url.URL) bool {
	// Skip links that could not be resolved
	if link == nil {
		return false
	}

	// Skip links with non http(s) schemes
	if link.Scheme != "http" && link.Scheme != "https" {
		return false
	}

	// Keep links to the same host
	if link.Hostname() == srcHost {
		return true
	}

	// Skip links that resolve to private networks
	if isPrivate := isPrivate(link.Host); isPrivate {
		return false
	}

	return true
}

// add "/" if link not end up with that
// ex: http://example.com/about become http://example.com/about/
func trailingSlash(s string) string {
	if s == "" {
		return "/"
	}

	if s[len(s)-1] != '/' {
		return s + "/"
	}
	return s
}

func resolveURL(relTo *url.URL, target string) *url.URL {
	tLen := len(target)
	if tLen == 0 {
		return nil
	}

	if tLen >= 1 && target[0] == '/' {
		if tLen >= 2 && target[1] == '/' {
			target = relTo.Scheme + ":" + target
		}
	}

	if targetURL, err := url.Parse(target); err == nil {
		return relTo.ResolveReference(targetURL)
	}

	return nil
}
