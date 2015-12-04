package gen

import (
	"errors"
	"fmt"
	"log"
)

type GenServer struct {
	done chan struct{}
	evts chan func()
}

func (gen *GenServer) loop() {
	for {
		select {
		case <-gen.done:
			return
		case evt := <-gen.evts:
			err := recoverWrapper(evt)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

var ErrServerNotRunning = errors.New("server not running")

func (gen *GenServer) Submit(fn func()) error {
	select {
	case gen.evts <- fn:
	default:
		return ErrServerNotRunning
	}
	return nil
}

func (gen *GenServer) SubmitChan(fn func()) <-chan error {
	ch := make(chan error, 1)
	ch <- gen.Submit(func() {
		defer close(ch)
		fn()
	})
	return ch
}

func (gen *GenServer) Stop() <-chan error {
	ch := make(chan error)
	err := <-gen.SubmitChan(func() {
		close(gen.done)
		close(ch)
	})
	if err != nil {
		close(ch)
	}
	return ch
}

func (gen *GenServer) Ping() <-chan error {
	ch := make(chan error)
	err := <-gen.SubmitChan(func() {
		close(ch)
	})
	if err != nil {
		close(ch)
	}
	return ch
}

func (gen *GenServer) Start() <-chan error {
	if gen == nil {
		gen = new(GenServer)
	}

	gen.done = make(chan struct{})

	if gen.evts == nil {
		gen.evts = make(chan func())
	}

	go gen.loop()
	return gen.Ping()
}

func recoverWrapper(fn func()) (err error) {
	defer func() {
		switch e := recover().(type) {
		case nil:
		case error:
			err = e
		default:
			err = fmt.Errorf("%v\n", e)
		}
	}()

	fn()
	return
}
