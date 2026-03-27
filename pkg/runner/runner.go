package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// Handler определяет компонент, который может быть запущен и остановлен.
type Handler interface {
	// Start запускает обработчик. Блокируется до тех пор, пока не произойдёт
	// ошибка или не будет отменён контекст.
	Start(context.Context) error
	// Stop останавливает обработчик. Вызывается после завершения Start.
	// Контекст в Stop может быть отдельным (например, с таймаутом).
	Stop(context.Context) error
}

// Runner управляет жизненным циклом Handler с обработкой сигналов ОС.
type Runner struct {
	handler     Handler
	signals     []os.Signal
	stopTimeout time.Duration // опциональный таймаут для Stop
	once        atomic.Bool   // защита от повторного запуска
}

// Option настраивает Runner.
type Option func(*Runner) error

// WithHandler устанавливает обрабатываемый компонент (обязательно).
func WithHandler(handler Handler) Option {
	return func(r *Runner) error {
		if handler == nil {
			return errors.New("runner: handler cannot be nil")
		}
		r.handler = handler
		return nil
	}
}

// WithSignals задаёт сигналы, которые будут приводить к остановке.
// Если не заданы, используются SIGINT и SIGTERM.
func WithSignals(signals ...os.Signal) Option {
	return func(r *Runner) error {
		r.signals = signals
		return nil
	}
}

// WithStopTimeout задаёт максимальное время ожидания остановки Handler.
// Если не задано, Stop будет использовать контекст, отменённый при завершении Start.
func WithStopTimeout(timeout time.Duration) Option {
	return func(r *Runner) error {
		r.stopTimeout = timeout
		return nil
	}
}

// New создаёт новый Runner с заданными опциями.
func New(opts ...Option) (*Runner, error) {
	r := &Runner{
		signals: []os.Signal{os.Interrupt, syscall.SIGTERM},
	}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	if r.handler == nil {
		return nil, errors.New("runner: handler is required")
	}
	return r, nil
}

// Run запускает обработчик и ожидает его завершения.
// Возвращает первую ошибку, возникшую в Start или Stop, либо nil при штатном
// завершении по сигналу. Метод не должен вызываться повторно.
func (r *Runner) Run(ctx context.Context) error {
	if !r.once.CompareAndSwap(false, true) {
		return errors.New("runner: run already called")
	}

	ctx, cancel := signal.NotifyContext(ctx, r.signals...)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("runner: panic in Start: %v", p)
				cancel()
			}
		}()
		return r.handler.Start(ctx)
	})

	g.Go(func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("runner: panic in Stop: %v", p)
				cancel()
			}
		}()
		<-ctx.Done()

		stopCtx := ctx
		if r.stopTimeout > 0 {
			var stopCancel context.CancelFunc
			stopCtx, stopCancel = context.WithTimeout(context.Background(), r.stopTimeout)
			defer stopCancel()
		}
		return r.handler.Stop(stopCtx)
	})

	return g.Wait()
}
