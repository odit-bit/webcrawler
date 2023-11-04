package crawler

import "regexp"

var exclude_Extensions = []string{".jpg", ".jpeg", ".png", ".gif", ".ico", ".css", ".js", ".pdf"}

var (
	baseHrefRegex = regexp.MustCompile(`(?i)<base.*?href\s*?=\s*?"(.*?)\s*?"`)

	findLinkRegex = regexp.MustCompile(`(?i)<a.*?href\s*?=\s*?"\s*?(.*?)\s*?".*?>`)

	exclusionRegex = regexp.MustCompile(`(?i)\.(?:jpg|jpeg|png|gif|ico|css|js|pdf|tar|zip|msi|gz|pkg)$`)

	// nofollowRegex = regexp.MustCompile(`(?i)rel\s*?=\s*?"?nofollow"?`)
)
