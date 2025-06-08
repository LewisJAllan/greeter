package asynchronous

import (
	"context"
	"sync"
)

// AsyncWaiter is a Listener that waits for asynchronous operations to finish before carrying on shutdown.
type AsyncWaiter struct {
	wg     sync.WaitGroup
	waitM  sync.Mutex
	waitCh chan struct{}
}

func NewAsyncWaiter() AsyncWaiter {
	return AsyncWaiter{}
}

func (w *AsyncWaiter) Wait() {
	w.wg.Wait()
}

func (w *AsyncWaiter) Add(delta int) {
	w.wg.Add(delta)
}

func (w *AsyncWaiter) Done() {
	w.wg.Done()
}

func (w *AsyncWaiter) Name() string {
	return "Wait for Asynchronous operations to complete"
}

func (w *AsyncWaiter) Start(_ context.Context) error {
	w.waitM.Lock()
	w.waitCh = make(chan struct{})
	w.waitM.Unlock()
	<-w.waitCh
	w.Wait()
	return nil
}

func (w *AsyncWaiter) Stop(_ context.Context) error {
	w.waitM.Lock()
	close(w.waitCh)
	w.waitM.Unlock()
	return nil
}

func (w *AsyncWaiter) Run(f func()) {
	w.Add(1)
	go func() {
		defer w.Done()
		f()
	}()
}
