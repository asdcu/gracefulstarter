package gracefulstarter

import (
	"os"
	"os/signal"
	"syscall"
)

// Start launch app with graceful shutdown
func Start(app App) error {
	return gracefulStart(app, InterruptCh)
}

// StartWithQuitCh start with quit func custom
func StartWithQuitCh(app App, interruptCh func() <-chan interface{}) error {
	return gracefulStart(app, interruptCh)
}

// interruptCh for unittest
func gracefulStart(app App, interruptCh func() <-chan interface{}) error {
	go func() {
		<-interruptCh()
		app.Stop()
	}()
	return app.Start()
}

// InterruptCh returns channel which will get data when system receives interrupt signal.
// stop worker with Ctrl+C.
func InterruptCh() <-chan interface{} {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ret := make(chan interface{}, 1)
	go func() {
		s := <-c
		ret <- s
		close(ret)
	}()

	return ret
}
