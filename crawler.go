package webcrawler

import (
	"context"

	"github.com/odit-bit/webcrawler/x/xpipe"
)

type Crawler struct {
	pipe *xpipe.Pipe[*Resource]
}

func NewCrawler() *Crawler {

	pipe := xpipe.New[*Resource](FetchHTML(), ExtractURLs(), ExtractHtmlContent())
	s := Crawler{
		pipe: pipe,
	}
	return &s
}

func (s *Crawler) Crawl(ctx context.Context, fetcher xpipe.Fetcher[*Resource], streamer xpipe.Streamer[*Resource]) error {

	crawlCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := s.pipe.Run(crawlCtx, fetcher, streamer)
	if err != nil {
		return err
	}

	return nil
}
