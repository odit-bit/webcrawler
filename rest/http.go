package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/odit-bit/webcrawler"
	"github.com/odit-bit/webcrawler/x/xpipe"
)

var _ http.Handler = (*api)(nil)

type api struct {
	crawler webcrawler.Crawler
}

func New() *api {
	a := api{
		crawler: *webcrawler.NewCrawler(),
	}
	return &a
}

// ServeHTTP implements http.Handler.
func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query().Get("url")
	if q == "" {
		http.Error(w, "url is nil", http.StatusBadRequest)
		return
	}

	resource := webcrawler.NewResource()
	resource.URL = q

	//we need datastructure to stream the request into crawler
	iter := requestPool.Get().(*requestIterator)
	defer iter.Close()
	iter.reqData = append(iter.reqData, resource)

	go func() {
		a.crawler.Crawl(r.Context(), iter, iter)
	}()

	// set as json to stream the result to response
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	f := w.(http.Flusher)
	for {
		select {
		case <-r.Context().Done():
		case res, ok := <-iter.resultC:
			if ok {
				res.Retrieved()
				err := enc.Encode(res)
				if err != nil {
					http.Error(w, "error encoding ", http.StatusInternalServerError)
					return
				} else {
					f.Flush()
					res.Put()
					continue
				}
			}

		}
		break
	}

}

var requestPool = sync.Pool{
	New: func() any {
		ri := &requestIterator{
			resultC: make(chan *webcrawler.Resource),
			reqData: []*webcrawler.Resource{},
			idx:     0,
		}

		return ri
	},
}

var _ xpipe.Fetcher[*webcrawler.Resource] = (*requestIterator)(nil)
var _ xpipe.Streamer[*webcrawler.Resource] = (*requestIterator)(nil)

type requestIterator struct {
	resultC chan *webcrawler.Resource
	reqData []*webcrawler.Resource
	idx     int
}

// Consume implements xpipe.Streamer.
func (ri *requestIterator) Consume(ctx context.Context, result <-chan *webcrawler.Resource) error {
	defer close(ri.resultC)
	for {
		select {
		case <-ctx.Done():
		case v, ok := <-result:
			if ok {
				select {
				case <-ctx.Done():
				case ri.resultC <- v:
					continue
				}
			}

		}
		break
	}

	return nil
}

// Error implements xpipe.Fetcher.
func (a *requestIterator) Error() error {
	return nil
}

// Next implements xpipe.Fetcher.
func (a *requestIterator) Next() bool {
	// log.Println("Api next data index :", a.idx)
	return a.idx < len(a.reqData)
}

// Resource implements xpipe.Fetcher.
func (a *requestIterator) Resource() *webcrawler.Resource {
	v := a.reqData[a.idx]
	a.idx++
	// log.Println("Api data fetched :", v.URL)
	return v
}

func (a *requestIterator) Close() {
	a.idx = 0
	a.reqData = a.reqData[:0]
	a.resultC = nil

	requestPool.Put(a)
}
