package xpipe

import (
	"context"
)

// Fetcher to fetch data from network
type Fetcher[T any] interface {
	Error() error
	Next() bool
	Resource() T
}

// Stream the result from pipe
type Streamer[T any] interface {
	Consume(ctx context.Context, result <-chan T) error
}

// Generic Pipe with FIFO approach
type Pipe[T any] struct {
	processors []ProcessorFunc[T]
}

// create New Pipe Instance with T type
func New[T any](processors ...ProcessorFunc[T]) *Pipe[T] {
	p := Pipe[T]{

		processors: processors,
	}

	return &p
}

// run the pipe with underlying process such producer and stage is executed concurrently,
// while consumer will acquire them all
// pipe will create stages according to supllied processors
func (p *Pipe[T]) Run(ctx context.Context, source Fetcher[T], sink Streamer[T]) error {

	producer := NewProducer[T](source)
	consumer := NewConsumer[T](sink)

	src := producer.Produce(ctx)

	var errChannels []<-chan error
	for _, processor := range p.processors {
		out, err := StageSemaphore[T](ctx, src, processor)
		src = out
		errChannels = append(errChannels, err)
	}

	//merge errors from stages into one channel
	errBus := merge(ctx, errChannels...)

	err := consumer.Consume(ctx, src, errBus)
	if err != nil {
		return err
	}
	return nil
}
