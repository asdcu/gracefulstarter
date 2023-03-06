package gracefulstarter

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

var _ App = (*empty)(nil)

var _ App = (*base)(nil)

type base struct {
	start func() error
	stop  func()
}

func (b base) Start() error { return b.start() }

func (b base) Stop() { b.stop() }

type empty struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (a *empty) Start() error {
	<-a.ctx.Done()
	return nil
}

func (a *empty) Stop() {
	a.cancel()
}

var _ App = (*array)(nil)

type array struct {
	root                *empty
	apps                []App
	eg                  *errgroup.Group
	egCtx               context.Context
	stopTimeoutDuration time.Duration
	once                sync.Once
}

func (a *array) Start() error {
	wg := &sync.WaitGroup{}
	for _, item := range a.apps {
		app := item
		a.eg.Go(func() error {
			<-a.egCtx.Done() // wait for stop signal
			app.Stop()
			return nil
		})
		wg.Add(1)
		a.eg.Go(func() error {
			wg.Done()
			defer a.once.Do(func() {
				a.root.Stop()
			})
			// non-blocking app start return will stop others app to stop
			return app.Start()
		})
	}
	// wait all app start
	wg.Wait()
	// start root and wait root stop
	return a.root.Start()
}

func (a *array) Stop() {
	// stop root app
	a.once.Do(func() {
		a.root.Stop()
	})
	// wait all apps stop in stop timeout
	a.stopWithTimeout(a.eg.Wait)
}

func (a *array) stopWithTimeout(stop func() error) {
	out := make(chan struct{}, 1)
	go func(fn func() error) {
		// ignore stop error. If necessary，Open via logger.Error function
		_ = fn()
		out <- struct{}{}
	}(stop)
	select {
	case <-out:
		close(out)
	case <-time.After(a.stopTimeoutDuration):
	}
}
