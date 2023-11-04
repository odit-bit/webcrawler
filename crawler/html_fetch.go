package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/odit-bit/webcrawler/x/xpipe"
)

func FetchHTML() xpipe.ProcessorFunc[*Resource] {

	return func(ctx context.Context, src *Resource) (*Resource, error) {
		pURL := src.URL
		// Skip URLs that point to files that cannot contain html content.
		if exclusionRegex.MatchString(pURL) {
			// log.Printf("link fetcher url: %v\nerror: %v \n", pURL, "not containt html content")
			return nil, fmt.Errorf("[FetchURL] error: not containt html page")
		}

		// Never crawl links in private networks (e.g. link-local addresses).
		// This is a security risk!
		ok := isPrivate(pURL)
		if ok {
			// log.Printf("link fetcher error: %v url: %v\n", err, pURL)
			return nil, fmt.Errorf("[FetchURL] error: %v is private address ", pURL)
		}

		//get url within timeout otherwise skipped
		if err := contentFromURL(ctx, src); err != nil {
			// log.Printf("link fetcher error: %v url: %v\n", err, pURL)
			return nil, err
		}

		return src, nil
	}
}

// get content of url pointed
func contentFromURL(ctx context.Context, payload *Resource) error {
	getter := http.DefaultClient

	// url Getter
	// held crawl link in expensive connection
	res, err := getter.Get(payload.URL)
	if err != nil {
		return fmt.Errorf("http request: %v", err)
	}

	// http.Response should not nil, skipped not success code
	if res == nil || res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("http response status nok ok (%v)", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "html") {
		return fmt.Errorf("http response: non html content-type:%v", contentType)
	}

	_, err = io.Copy(&payload.rawBuffer, res.Body)
	if err != nil {
		return fmt.Errorf("copy response body: %v", err)
	}

	err = res.Body.Close()
	if err != nil {
		return fmt.Errorf("close response body: %v", err)
	}

	// log.Println("link fetcher content type", contentType, "url:", payload.URL)
	return nil
}
