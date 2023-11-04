package crawler

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/odit-bit/webcrawler/x/xpipe"
)

// ExtractURLs only extract link that has the same host as the domain in html page

// Process implements pipeline.Processor.
func ExtractURLs() xpipe.ProcessorFunc[*Resource] {

	return func(ctx context.Context, src *Resource) (*Resource, error) {
		payload := src
		relTo, err := url.Parse((payload.URL))
		if err != nil {
			return nil, err
		}

		//search <base> tag n resolve to absolute url
		// <base href="XXX">
		content := payload.rawBuffer.String()
		baseMatch := baseHrefRegex.FindStringSubmatch(content)
		if len(baseMatch) == 2 {
			base := resolveURL(relTo, baseMatch[1])
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

			var ok bool
			link, ok = isValidate(relTo, link)
			if !ok {
				continue
			}

			linkStr := link.String()
			if _, seen := seenMap[linkStr]; seen {
				continue
			}

			seenMap[linkStr] = struct{}{}
			payload.FoundURLs = append(payload.FoundURLs, linkStr)
		}

		return payload, nil

	}
}

func isContainNonHTML(link string) bool {
	ext := path.Ext(link)
	switch ext {
	case ".html", "htm", "":
		return false
	default:
		return true
	}
}

func isValidate(relTo *url.URL, link *url.URL) (*url.URL, bool) {
	// Skip links that could not be resolved
	if link == nil {
		return nil, false
	}

	link.Path = strings.TrimRight(link.Path, "/")

	// Skip links to the other host
	if link.Host != relTo.Host {
		return nil, false
	}

	// Skip links with non http(s) schemes
	if link.Scheme != "http" && link.Scheme != "https" {
		return nil, false
	}

	if isContainNonHTML(link.Path) {
		return nil, false
	}

	link.Path = strings.TrimSuffix(link.Path, ".html")
	link.Fragment = ""

	// // Skip links that resolve to private networks
	// if isPrivate := isPrivate(link.Host); isPrivate {
	// 	return false
	// }

	return link, true
}

// add "/" if link not end up with that
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
