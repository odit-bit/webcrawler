package webcrawler

import (
	"context"
	"log"
	"time"

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
	start := time.Now()
	crawlCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := s.pipe.Run(crawlCtx, fetcher, streamer)
	if err != nil {
		return err
	}

	log.Printf("crawled finish %v", time.Since(start).Round(time.Second))
	return nil
}
