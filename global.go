package shutdown

import (
	"context"
	"time"
)

var (
	global = New()
)

func SetLogger(l Logger) {
	global.log = l
}

func SetGracefulTimeout(timeout time.Duration) {
	global.gracefulTimeout = timeout
}

func AddGraceful(ff ...func(ctx context.Context)) {
	global.AddGraceful(ff...)
}

func Done() error {
	return global.Done()
}

func Cancel(err error) {
	global.Cancel(err)
}

func Listen() {
	global.Listen()
}
