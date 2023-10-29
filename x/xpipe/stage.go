package xpipe

import (
	"context"
	"log"
	"sync"

	"golang.org/x/sync/semaphore"
)

type ProcessorFunc[T any] func(ctx context.Context, src T) (T, error)

// run data processing concurrently
func Stage[T any](
	ctx context.Context,
	src <-chan T,
	process ProcessorFunc[T],
) (<-chan T, <-chan error) {

	resC := make(chan T)
	errC := make(chan error)

	go func() {
		defer close(resC)
		defer close(errC)

		for {
			select {
			case <-ctx.Done():
				return
			case rsc, ok := <-src:
				if !ok {
					// maybe close
					return
				}

				result, err := process(ctx, rsc)

				if err != nil {
					select {
					case <-ctx.Done():
						return
					case errC <- err:
						continue
					}
				}

				// prevent blocking if jobs is done
				select {
				case <-ctx.Done():
					return
				case resC <- result:
				}
			}
		}

	}()

	return resC, errC
}

// run data processing concurrently use semaphore to utilize all cpu available
func StageSemaphore[T any](
	ctx context.Context,
	src <-chan T,
	process ProcessorFunc[T],
) (<-chan T, <-chan error) {

	resC := make(chan T)
	errC := make(chan error)

	limit := int64(2)
	// Use all CPU cores to maximize efficiency. We'll set the limit to 2 so you
	// can see the values being processed in batches of 2 at a time, in parallel
	// limit := int64(runtime.NumCPU())
	sem1 := semaphore.NewWeighted(limit)

	go func() {
		defer close(resC)
		defer close(errC)

		for {
			select {
			case <-ctx.Done():
				return
			case rsc, ok := <-src:
				if !ok {
					// maybe close
					if err := sem1.Acquire(ctx, limit); err != nil {
						log.Printf("Failed to acquire semaphore: %v", err)
					}
					return
				}

				// Acquire a semaphore
				if err := sem1.Acquire(ctx, 1); err != nil {
					log.Println("error acquire cpu", err)
					break
				}

				// go routine process
				go func() {
					//release semaphore
					defer sem1.Release(1)
					result, err := process(ctx, rsc)

					if err != nil {
						select {
						case <-ctx.Done():
							return
						case errC <- err:
						}
					} else {
						select {
						case <-ctx.Done():
							return
						case resC <- result:
						}
					}
				}()
			}
		}

	}()

	return resC, errC
}

// merge is generic FAN-IN implementation
func merge[T any](ctx context.Context, ch ...<-chan T) <-chan T {
	var wg sync.WaitGroup

	out := make(chan T)

	//merge ch to out
	//out chan will receive any c chan that ready (not empty)
	//
	for _, c := range ch {
		wg.Add(1)
		go func(c <-chan T) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case n, ok := <-c:
					if !ok {
						// maybe close
						return
					}

					out <- n
				}
			}
		}(c)
	}

	// wait all goroutines exit
	go func() {
		defer func() {
			log.Println("exit merge bus")
		}()
		defer close(out)
		wg.Wait()

	}()

	return out
}
