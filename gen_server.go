package gen

import (
	"fmt"
	"log"
)

type GenServer struct {
	evts chan func()
}

func (gen *GenServer) loop() {
	for evt := range gen.evts {
		err := recoverWrapper(evt)
		if err != nil {
			log.Println(err)
		}
	}
}

func (gen *GenServer) Submit(fn func()) {
	gen.evts <- fn
}

func (gen *GenServer) Stop() <-chan struct{} {
	ch := make(chan struct{})
	gen.evts <- func() {
		close(gen.evts)
		close(ch)
	}
	return ch
}

func (gen *GenServer) Ping() <-chan struct{} {
	ch := make(chan struct{})
	gen.evts <- func() {
		close(ch)
	}
	return ch
}

func (gen *GenServer) Start() <-chan struct{} {
	if gen == nil {
		gen = new(GenServer)
	}
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
