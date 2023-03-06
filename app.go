package gracefulstarter

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

// App minimum runnable unit
type App interface {
	// Start launch app and blocking it
	Start() error
	// Stop notify to closing start func, even need to wait for the startup function to exit completely
	Stop()
}

func Empty() App {
	ctx, cancel := context.WithCancel(context.Background())
	return &empty{ctx: ctx, cancel: cancel}
}

func Functional(start func() error, stop func()) App {
	return &base{start: start, stop: stop}
}

func Array(apps []App) App {
	app := Empty()
	if len(apps) == 0 {
		return app
	}
	root := app.(*empty)
	eg, egCtx := errgroup.WithContext(root.ctx)
	//  default 30 seconds quit
	return &array{apps: apps, egCtx: egCtx, eg: eg, root: root, stopTimeoutDuration: 30 * time.Second}
}
