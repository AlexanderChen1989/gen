package gen

import (
	"fmt"
	"testing"
	"time"
)

type PanicServer struct {
	GenServer
}

func TestSuperviser(t *testing.T) {
	sup := newSuperviser(nil)
	var ps *PanicServer
	sup.Add(func(server GenServer) (GenServer, error) {
		fmt.Println("create new PanicServer...")
		ps = &PanicServer{server}
		return ps, nil
	})
	sup.Start()

	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		ps.Stop()
	}
}
