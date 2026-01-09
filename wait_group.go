package goshutdown

import (
	"context"
	"sync"
)

type WaitGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	sync.WaitGroup
}

func NewWaitGroup() *WaitGroup {
	ctx, cancel := context.WithCancel(context.Background())
	return &WaitGroup{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (wg *WaitGroup) Wait() {
	defer wg.cancel()
	wg.WaitGroup.Wait()
}
