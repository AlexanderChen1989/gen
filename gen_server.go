package gen

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
)

type taskType int

const (
	syncTask taskType = iota
	asyncTask
)

type task struct {
	typ     taskType
	timeout time.Duration
	action  func(ctx context.Context)
}

func (t task) Do(ctx context.Context) {
	if t.timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, t.timeout)
	}
	switch t.typ {
	case syncTask:
		t.action(ctx)
	case asyncTask:
		go t.action(ctx)
	}
}

func recoverTask(t task) task {
	orig := t.action

	t.action = func(ctx context.Context) {
		defer func() {
			recover()
		}()

		orig(ctx)
	}

	return t
}

func Async(t task) task {
	t.typ = asyncTask
	return t
}

func Timeout(timeout time.Duration) func(task) task {
	return func(t task) task {
		t.timeout = timeout
		return t
	}
}

type GenServer interface {
	Start() error
	Stop() error
	Ping() error
	Submit(fn func(ctx context.Context), setups ...func(task) task) error
}

func New(ctx context.Context) GenServer {
	return newServer(ctx)
}

func newServer(ori context.Context) *genServer {
	ctx, cancel := context.WithCancel(ori)
	return &genServer{
		ctx:   ctx,
		ping:  make(chan struct{}),
		tasks: make(chan task),
		cancel: func() {
			fmt.Println("Cancel")
			cancel()
		},
	}
}

type genServer struct {
	ctx        context.Context
	cancel     func()
	cancelOnce sync.Once
	startOnce  sync.Once
	ping       chan struct{}
	tasks      chan task
}

func (s *genServer) loop() {
	fmt.Println("Start")
	for {
		select {
		case t := <-s.tasks:
			t.Do(s.ctx)
		case s.ping <- struct{}{}:
		case <-s.ctx.Done():
			fmt.Println("Done")
			return
		}
	}
}

var ErrNotRunning = errors.New("genServer not running")

func (s *genServer) submit(t task, setups ...func(task) task) error {
	for _, setup := range setups {
		t = setup(t)
	}
	select {
	case s.tasks <- t:
		return nil
	case <-s.ctx.Done():
		return ErrNotRunning
	}
}

func (s *genServer) Submit(fn func(ctx context.Context), setups ...func(task) task) error {
	return s.submit(task{typ: syncTask, action: fn}, append(setups, recoverTask)...)
}

func (s *genServer) Start() error {
	go s.startOnce.Do(s.loop)
	return nil
}

func (s *genServer) Stop() error {
	s.cancelOnce.Do(s.cancel)
	return nil
}

var ErrTimeout = errors.New("timeout")

func (s *genServer) Ping() error {
	select {
	case <-s.ping:
		return nil
	case <-time.After(time.Second):
		return ErrTimeout
	}
}
