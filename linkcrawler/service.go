package linkcrawler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/linkstore/linkgraph"
	"github.com/odit-bit/webcrawler"
	"github.com/odit-bit/webcrawler/x/xpipe"
)

var minUUID = uuid.Nil
var maxUUID = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
var default_interval = 1 * time.Minute
var default_recrawl_interval = 7 * 24 * time.Hour

type CrawlService struct {
	crawler *webcrawler.Crawler
	api     linkgraph.Graph

	Interval        time.Duration //default 1 * time.Minute
	RecrawlInterval time.Duration
}

func New(linkAPI linkgraph.Graph) *CrawlService {
	s := CrawlService{
		crawler: webcrawler.NewCrawler(),
		api:     linkAPI,

		Interval:        default_interval,
		RecrawlInterval: default_recrawl_interval,
	}
	return &s
}

// return fetcher to supply data for pipe
func (li *CrawlService) Fetch(fromID, toID uuid.UUID, retrieveBefore time.Time) (*linkFetcher, error) {
	iter, err := li.api.Links(fromID, toID, retrieveBefore)
	if err != nil {
		return nil, err
	}

	fetcher := &linkFetcher{
		rpc: iter,
	}

	return fetcher, nil
}

// return streamer to send data from pipe
func (li *CrawlService) Dispatch() (*linkDispatcher, error) {
	consumer := &linkDispatcher{
		api: li.api,
	}

	return consumer, nil
}

var _ xpipe.Fetcher[*webcrawler.Resource] = (*linkFetcher)(nil)

type linkFetcher struct {
	rpc linkgraph.LinkIterator
}

// Error implements xpipe.Fetcher.
func (lf *linkFetcher) Error() error {
	return lf.rpc.Error()
}

// Next implements xpipe.Fetcher.
func (lf *linkFetcher) Next() bool {
	return lf.rpc.Next()

}

// Resource implements xpipe.Fetcher.
func (lf *linkFetcher) Resource() *webcrawler.Resource {
	l := lf.rpc.Link()
	resource := webcrawler.NewResource()
	resource.ID = l.ID
	resource.URL = l.URL

	return resource
}

func (lf *linkFetcher) Close() {
	lf.rpc.Close()
}

var _ xpipe.Streamer[*webcrawler.Resource] = (*linkDispatcher)(nil)

type linkDispatcher struct {
	api linkgraph.Graph
}

// Consume implements xpipe.Streamer.
func (ld *linkDispatcher) Consume(ctx context.Context, result <-chan *webcrawler.Resource) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case r, ok := <-result:
			if !ok {
				return nil
			}
			err := ld.upsertResource(r)
			if err != nil {
				return err
			}
			r.Put()
		}
	}
}

func (ld *linkDispatcher) upsertResource(r *webcrawler.Resource) error {
	link := &linkgraph.Link{
		ID:          r.ID,
		URL:         r.URL,
		RetrievedAt: time.Now(),
	}
	if err := ld.api.UpsertLink(link); err != nil {
		return err
	}

	for _, dst := range r.FoundURLs {
		dstLink := &linkgraph.Link{
			URL: dst,
		}

		//insert link destination as node
		err := ld.api.UpsertLink(dstLink)
		if err != nil {
			return err
		}

		//insert link destination as edge
		edge := linkgraph.Edge{

			Src: link.ID,
			Dst: dstLink.ID,
		}

		if err := ld.api.UpsertEdge(&edge); err != nil {
			return err
		}

	}

	removeEdgeBefore := time.Now()
	if err := ld.api.RemoveStaleEdges(link.ID, removeEdgeBefore); err != nil {
		return err
	}

	return nil
}

func (la *CrawlService) Run(ctx context.Context) error {
	ticker := time.NewTicker(la.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := la.startCrawl(ctx)
			if err != nil {
				return err
			}
			ticker.Reset(la.Interval)
		}
	}
}

//underlying crawler need fetcher and streamer for source and sink, that this type profided
//from api

func (la *CrawlService) startCrawl(ctx context.Context) error {

	fetch, err := la.Fetch(minUUID, maxUUID, time.Now().Add(-la.RecrawlInterval))
	if err != nil {
		return err
	}
	dispatch, _ := la.Dispatch()

	if err := la.crawler.Crawl(ctx, fetch, dispatch); err != nil {
		return err
	}
	fetch.Close()

	return nil
}
