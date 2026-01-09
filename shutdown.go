package goshutdown

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Logger interface {
	Info(msg string, args ...any)
}

type Shutdown struct {
	log             Logger
	exit            chan os.Signal
	gracefulFuncs   []func(ctx context.Context)
	gracefulTimeout time.Duration

	ctx       context.Context
	ctxCancel context.CancelCauseFunc
}

func New() *Shutdown {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &Shutdown{
		log:             slog.Default(),
		exit:            make(chan os.Signal, 1),
		gracefulFuncs:   []func(ctx context.Context){},
		gracefulTimeout: 5 * time.Second,
		ctx:             ctx,
		ctxCancel:       cancel,
	}
}

func (s *Shutdown) SetLogger(l Logger) {
	s.log = l
}

func (s *Shutdown) SetGracefulTimeout(timeout time.Duration) {
	s.gracefulTimeout = timeout
}

func (s *Shutdown) AddGraceful(ff ...func(ctx context.Context)) {
	s.gracefulFuncs = append(s.gracefulFuncs, ff...)
}

func (s *Shutdown) Cancel(err error) {
	s.ctxCancel(err)
}

func (s *Shutdown) Done() error {
	<-s.ctx.Done()
	err := context.Cause(s.ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func (s *Shutdown) waitGraceful() {
	if len(s.gracefulFuncs) == 0 {
		s.log.Info("Shutdown no graceful functions")
		return
	}

	waitCtx, waitCancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
	defer waitCancel()

	wg := NewWaitGroup()
	for _, f := range s.gracefulFuncs {
		wg.Go(func() {
			f(waitCtx)
		})
	}

	go wg.Wait()

	select {
	case <-waitCtx.Done():
	case <-wg.ctx.Done():
	}

}

func (s *Shutdown) Listen() {
	go func() {
		defer s.Cancel(nil)
		signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
		signal.Notify(s.exit, signals...)
		s.log.Info("Shutdown listen", "signals", signals)

		sig := <-s.exit
		s.log.Info("Shutdown signal received", "signal", sig)

		s.waitGraceful()
	}()
}
