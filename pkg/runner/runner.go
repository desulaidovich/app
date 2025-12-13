package runner

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
)

type Handler interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type Runner struct {
	handler Handler
	wg      sync.WaitGroup
}

func New(handler Handler) *Runner {
	return &Runner{
		handler: handler,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 2)
	done := make(chan struct{})

	r.wg.Go(func() {
		if err := r.handler.Start(ctx); err != nil {
			select {
			case errChan <- err:
				stop()

			default:
			}
		}
	})

	r.wg.Go(func() {
		<-ctx.Done()

		if err := r.handler.Stop(ctx); err != nil {
			select {
			case errChan <- err:
				stop()

			default:
			}
		}
	})

	go func() {
		r.wg.Wait()
		close(done)
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err

	case <-done:
		return nil
	}
}
