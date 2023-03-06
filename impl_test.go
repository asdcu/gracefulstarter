package gracefulstarter

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_array(t *testing.T) {
	t.Run("graceful start and stop", func(t *testing.T) {
		p1 := newPrinter(t, 2*time.Second, "NO.1")
		p2 := newPrinter(t, 5*time.Second, "NO.2")
		mul := Array([]App{p1, p2})
		time.AfterFunc(8300*time.Millisecond, mul.Stop)
		assert.NoError(t, mul.Start())
		assert.EqualValues(t, 4, p1.cnt)
		assert.EqualValues(t, 1, p2.cnt)
	})
	t.Run("invalid stop", func(t *testing.T) {
		p1 := newPrinter(t, 2*time.Second, "NO.1")
		p2 := newPrinter(t, 5*time.Second, "NO.2")
		mul := Array([]App{p1, &invalidStopper{p2}})
		time.AfterFunc(8300*time.Millisecond, mul.Stop)
		assert.NoError(t, mul.Start())
		assert.EqualValues(t, 4, p1.cnt)
	})
	t.Run("non-blocking start", func(t *testing.T) {
		p1 := newPrinter(t, 2*time.Second, "NO.1")
		p2 := newPrinter(t, 5*time.Second, "NO.2")
		mul := Array([]App{p1, &invalidStarter{p2}})
		time.AfterFunc(8300*time.Millisecond, mul.Stop)
		assert.NoError(t, mul.Start())
		assert.EqualValues(t, 0, p1.cnt)
	})
}

var _ App = (*printer)(nil)

type printer struct {
	t      *testing.T
	ticker *time.Ticker
	msg    string
	cnt    int
}

func newPrinter(t *testing.T, d time.Duration, msg string) *printer {
	return &printer{t: t, ticker: time.NewTicker(d), msg: msg}
}

func (p *printer) Start() error {
	for cur := range p.ticker.C {
		p.t.Log(cur, ":", p.msg)
		p.cnt += 1
	}
	return nil
}

func (p *printer) Stop() {
	p.t.Log("stop:", p.msg)
	p.ticker.Stop()
}

type invalidStopper struct {
	*printer
}

func (i *invalidStopper) Stop() {
	for {
		i.t.Log("infinite loop")
		time.Sleep(time.Second)
	}
}

type invalidStarter struct {
	*printer
}

func (i *invalidStarter) Start() error {
	i.t.Log("non-blocking start")
	return nil
}

type startWithError struct {
	*printer
}

func (i *startWithError) Start() error {
	return errors.New("start with error")
}
