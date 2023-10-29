package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/odit-bit/webcrawler"
	"github.com/odit-bit/webcrawler/x/xpipe"
)

var _ http.Handler = (*api)(nil)

type api struct{}

func New() *api {
	a := api{}
	return &a
}

// ServeHTTP implements http.Handler.
func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var reqData []webcrawler.Resource
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//we need datastructure to stream the request into crawler
	iter := &requestIterator{
		resultC: make(chan *webcrawler.Resource),
		reqData: reqData,
		idx:     0,
	}

	cl := webcrawler.NewCrawler(iter, iter)

	go func() {
		cl.Crawl(r.Context())
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
					continue
				}
			}

		}
		break
	}

}

var _ xpipe.Fetcher[*webcrawler.Resource] = (*requestIterator)(nil)
var _ xpipe.Streamer[*webcrawler.Resource] = (*requestIterator)(nil)

type requestIterator struct {
	resultC chan *webcrawler.Resource
	reqData []webcrawler.Resource
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
	return &v
}
