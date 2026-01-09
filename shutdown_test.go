package shutdown

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"syscall"
	"testing"
	"time"
)

func equal(t *testing.T, a, b any) {
	t.Helper()
	if a != b {
		t.Errorf("%v != %v\n", a, b)
	}
}

func notEqual(t *testing.T, a, b any) {
	t.Helper()
	if a == b {
		t.Errorf("%v == %v\n", a, b)
	}
}

func TestSetLogger(t *testing.T) {
	newLog := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	shutdown := New()
	shutdown.SetLogger(newLog)
	equal(t, shutdown.log, newLog)
}

func TestSetGracefulTimeout(t *testing.T) {
	timeout := 3 * time.Second

	shutdown := New()
	shutdown.SetGracefulTimeout(timeout)
	equal(t, shutdown.gracefulTimeout, timeout)
}

func TestSigInt(t *testing.T) {
	shutdown := New()
	shutdown.Listen()

	go func() {
		time.Sleep(1 * time.Second)
		shutdown.exit <- syscall.SIGINT
	}()

	err := shutdown.Done()
	equal(t, err, nil)
}

func TestCancel(t *testing.T) {
	shutdown := New()
	shutdown.Listen()
	testErr := fmt.Errorf("test err")
	go func() {
		time.Sleep(1 * time.Second)
		shutdown.Cancel(testErr)
	}()

	err := shutdown.Done()
	equal(t, testErr, err)
}

func TestWaitGracefulStuck(t *testing.T) {
	shutdown := New()
	shutdown.AddGraceful(func(context.Context) {
		time.Sleep(60 * time.Second)
	})

	start := time.Now()
	shutdown.waitGraceful()
	equal(t, time.Since(start) < 2*shutdown.gracefulTimeout, true)
}

func TestGracefulFunc(t *testing.T) {
	shutdown := New()

	gracefulCtx1, gracefulCancel1 := context.WithCancel(context.Background())
	shutdown.AddGraceful(func(context.Context) {
		time.Sleep(4 * time.Second)
		gracefulCancel1()
	})

	gracefulCtx2, gracefulCancel2 := context.WithCancel(context.Background())
	shutdown.AddGraceful(func(context.Context) {
		time.Sleep(4 * time.Second)
		gracefulCancel2()
	})

	shutdown.Listen()
	go func() {
		time.Sleep(1 * time.Second)
		shutdown.exit <- syscall.SIGINT
	}()

	equal(t, nil, shutdown.Done())
	time.Sleep(1 * time.Second)

	notEqual(t, gracefulCtx1.Err(), nil)
	notEqual(t, gracefulCtx2.Err(), nil)
}
