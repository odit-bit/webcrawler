package crawler

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/odit-bit/webcrawler/x/xpipe"
)

var _ xpipe.Fetcher[*Resource] = (*mockSource)(nil)

type mockSource struct {
	ds  []*Resource
	idx int
}

// Error implements LinkSource.
func (*mockSource) Error() error {
	return nil
}

// Next implements LinkSource.
func (ms *mockSource) Next() bool {
	return ms.idx < len(ms.ds)
}

// Resource implements LinkSource.
func (ms *mockSource) Resource() *Resource {
	res := ms.ds[ms.idx]
	ms.idx++
	return res
}

var _ xpipe.Streamer[*Resource] = (*mockSink)(nil)

type mockSink struct{}

// Consume implements xpipe.Streamer.
func (ms *mockSink) Consume(ctx context.Context, result <-chan *Resource) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v, ok := <-result:
			if !ok {
				return ctx.Err()
			} else {
				fmt.Println("[crawle link] ", v)
			}
		}
	}
}

func Test_crawl(t *testing.T) {
	// // data source instance
	source := mockSource{
		ds: []*Resource{
			{
				ID:  uuid.New(),
				URL: "https://go.dev",
			},
		},
		idx: 0,
	}

	// data sink function
	sink := mockSink{}

	crawler := NewCrawler()
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err := crawler.Crawl(ctx, &source, &sink); err != nil {
		t.Fatal(err)
	}

}
