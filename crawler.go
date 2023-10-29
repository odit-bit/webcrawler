package webcrawler

import (
	"context"

	"github.com/odit-bit/webcrawler/x/xpipe"
)

type Crawler struct {
	pipe *xpipe.Pipe[*Resource]
}

func NewCrawler(fetcher xpipe.Fetcher[*Resource], streamer xpipe.Streamer[*Resource]) *Crawler {
	// data source instance
	producer := xpipe.NewProducer(fetcher)
	consumer := xpipe.NewConsumer(streamer)

	pipe := xpipe.New[*Resource](producer, consumer)
	s := Crawler{
		pipe: pipe,
	}
	return &s
}

func (s *Crawler) Crawl(ctx context.Context) error {

	crawlCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	
	err := s.pipe.Run(crawlCtx, FetchHTML(), ExtractURLs(), ExtractHtmlContent())
	if err != nil {
		return err
	}

	return nil
}
