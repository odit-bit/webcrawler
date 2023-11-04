package crawler

import (
	"context"

	"github.com/odit-bit/webcrawler/x/xpipe"
)

//implementation of xpipe pipeline with Resource as concrete type
//this package also defined the processors that use by stages

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
