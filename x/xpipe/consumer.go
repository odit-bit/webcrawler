package xpipe

import (
	"context"
	"log"
	"sync"
)

type ConsumerFunc[T any] func(ctx context.Context, result <-chan T) error

// Generic consumer instance
type Consumer[T any] struct {
	stream Streamer[T]
}

func NewConsumer[T any](streamer Streamer[T]) *Consumer[T] {
	c := Consumer[T]{
		stream: streamer,
	}
	return &c
}

// consume data-bus , and error-bus from pipe
func (c *Consumer[T]) Consume(ctx context.Context, resultC <-chan T, errBus <-chan error) error {
	// defer func() {
	// 	log.Println("exit Consume")
	// }()

	var wg sync.WaitGroup
	consumeC := make(chan T)

	//consume err from err-buss
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errBus:
				if !ok {
					// maybe close
					return
				}
				log.Println("consumer pipe :", err)
			}
		}
	}()

	// consume result from last stage (data buss)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(consumeC)

		for {
			select {
			case v, ok := <-resultC:
				if !ok {
					//channel maybe close
					// log.Println("pipe consumer: result chan is closed")
					return
				}
				select {
				case <-ctx.Done():
					return
				case consumeC <- v:
					// log.Println("pipe consumer: try send result")
				}

			case <-ctx.Done():

			}

		}
	}()

	err := c.stream.Consume(ctx, consumeC)
	wg.Wait()
	return err
}
