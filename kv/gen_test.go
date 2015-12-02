package kv

import (
	"fmt"
	"log"
	"testing"
)

type GenServer struct {
	kv   map[string]string
	evts chan func()
}

func (gs *GenServer) Put(k, v string) {
	gs.evts <- func() {
		gs.kv[k] = v
	}
}

func (gs *GenServer) Get(k string) string {
	ch := make(chan string)
	gs.evts <- func() {
		ch <- gs.kv[k]
	}
	return <-ch
}

func (gs *GenServer) Panic(msg string) {
	gs.evts <- func() {
		panic(msg)
	}
}

func recoverWrapper(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v\n", e)
		}
	}()
	fn()
	return
}

func (gs *GenServer) Wait() <-chan struct{} {
	ch := make(chan struct{})
	gs.evts <- func() {
		close(ch)
	}
	return ch
}

func (gs *GenServer) loop() {
	for evt := range gs.evts {
		if err := recoverWrapper(evt); err != nil {
			log.Println(err)
		}
	}
}

func TestGenServer(t *testing.T) {
	gs := &GenServer{
		kv:   make(map[string]string),
		evts: make(chan func()),
	}
	go gs.loop()
	gs.Panic("Hello1")
	gs.Panic("Hello2")
	gs.Panic("Hello3")
	gs.Put("Hello", "Word")
	fmt.Println(gs.Get("Hello"))
	<-gs.Wait()
}
