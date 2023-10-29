package xpipe

import (
	"context"
	"log"
)

type Producer[T any] struct {
	fetcher Fetcher[T]
}

func NewProducer[T any](fetcher Fetcher[T]) *Producer[T] {
	p := Producer[T]{
		fetcher: fetcher,
	}

	return &p
}

// return chan that supply data from source
func (p *Producer[T]) Produce(ctx context.Context) <-chan T {
	rscC := make(chan T)

	go func() {
		//close chan as soon as goroutine exit
		defer func() {
			close(rscC)
			log.Println("exit produce")
		}()

		//loop the source
		for p.fetcher.Next() {
			r := p.fetcher.Resource()
			select {
			case rscC <- r:
			case <-ctx.Done():
				return
			}
		}

		if err := p.fetcher.Error(); err != nil {
			log.Println("link source error", err)
			return
		}

	}()

	return rscC
}
