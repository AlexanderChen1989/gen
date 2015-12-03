package gen

import (
	"fmt"
	"sync"
	"testing"
)

type KVMServer struct {
	GenServer
	kv map[string]string
}

func (gs *KVMServer) Put(k, v string) {
	gs.Submit(func() {
		gs.kv[k] = v
	})
}

func (gs *KVMServer) Get(k string) string {
	ch := make(chan string)
	gs.Submit(func() {
		ch <- gs.kv[k]
	})
	return <-ch
}

type Task interface {
	Start() <-chan struct{}
	Stop() <-chan struct{}
	Ping() <-chan struct{}
}

type Supervisor struct {
	GenServer
	tasks []Task
}

func (sup *Supervisor) Stop() <-chan struct{} {
	ch := make(chan struct{})

	sup.Submit(func() {
		// stop all tasks
		var wg sync.WaitGroup
		wg.Add(len(sup.tasks))
		for _, task := range sup.tasks {
			go func(t Task) {
				<-t.Stop()
				wg.Done()
			}(task)
		}
		wg.Wait()

		// stop self
		<-sup.GenServer.Stop()
		close(ch)
	})

	return ch
}

func (sup *Supervisor) AddTask(tasks ...Task) <-chan struct{} {
	ch := make(chan struct{})

	sup.Submit(func() {
		sup.tasks = append(sup.tasks, tasks...)

		// start all tasks
		var wg sync.WaitGroup
		wg.Add(len(tasks))
		for _, task := range sup.tasks {
			go func(t Task) {
				<-t.Start()
				wg.Done()
			}(task)
		}
		wg.Wait()

		close(ch)
	})

	return ch
}

func TestSupervisor(t *testing.T) {
	kvm := &KVMServer{
		kv: make(map[string]string),
	}
	sup := new(Supervisor)
	sup.Start()
	<-sup.Start()
	<-sup.AddTask(kvm)
	kvm.Put("Hello", "100")
	fmt.Println(kvm.Get("Hello"))
	<-sup.Stop()
}
