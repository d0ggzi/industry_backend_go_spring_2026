package main

import (
	"context"
	"errors"
	"sync"
)

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {
	if workers < 1 {
		return nil, errors.New("workers count must be >= 1")
	}
	if len(in) == 0 {
		return []R{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make([]R, len(in))
	indices := make(chan int)

	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case idx, ok := <-indices:
					if !ok {
						return
					}

					result, err := fn(ctx, in[idx])
					if err != nil {
						errOnce.Do(func() {
							firstErr = err
							cancel()
						})
						return
					}
					out[idx] = result
				}
			}
		}()
	}

	go func() {
		defer close(indices)
		for i := range in {
			select {
			case <-ctx.Done():
				return
			case indices <- i:
			}
		}
	}()

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return out, nil
}
