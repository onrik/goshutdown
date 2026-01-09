package goshutdown

import "testing"

func TestWaitGroup(t *testing.T) {
	wg := NewWaitGroup()

	wg.Wait()
	err := wg.ctx.Err()
	if err == nil {
		t.Error("ctx not done")
	}
}
