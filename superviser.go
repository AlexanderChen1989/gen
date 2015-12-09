package gen

import (
	"time"

	"golang.org/x/net/context"
)

type server struct {
	gen     func(GenServer) (GenServer, error)
	server  GenServer
	started bool
	err     error
}

type Superviser struct {
	*genServer
	servers []*server
}

func newSuperviser(ctx context.Context) *Superviser {
	return &Superviser{
		genServer: newServer(ctx),
	}
}

func (sup *Superviser) each(fn func(*server)) {
	for _, s := range sup.servers {
		fn(s)
	}
}

func (sup *Superviser) Each(fn func(server GenServer, started bool, err error)) {
	sup.each(func(s *server) {
		fn(s.server, s.started, s.err)
	})
}

func (sup *Superviser) Add(children ...func(GenServer) (GenServer, error)) {
	for _, child := range children {
		s := &server{
			gen: child,
		}
		sup.servers = append(sup.servers, s)
	}
}

func (sup *Superviser) pingChild(s *server) {
	if err := s.server.Ping(20 * time.Millisecond); err != nil {
		// restart server
		s.server.Stop()
		s.started = false
		s.server, s.err = s.gen(New(sup.ctx))
		if s.err != nil {
			return
		}
		s.err = s.server.Start()
		if s.err != nil {
			return
		}
		s.started = true
	}
}

// ping started server
func (sup *Superviser) pingLoop() {
	for {
		select {
		case <-sup.ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			for _, s := range sup.servers {
				if !s.started {
					continue
				}
				go sup.pingChild(s)
			}
		}
	}
}

func (sup *Superviser) Start() (err error) {
	if err = sup.genServer.Start(); err != nil {
		return
	}

	sup.each(func(s *server) {
		s.server, s.err = s.gen(newServer(sup.ctx))
		if s.err != nil {
			err = s.err
			return
		}
		s.err = s.server.Start()
		if s.err != nil {
			err = s.err
			return
		}
		s.started = true
	})

	go sup.pingLoop()

	return
}
