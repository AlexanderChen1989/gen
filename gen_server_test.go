package gen

import (
	"fmt"
	"sync"
	"testing"
	"time"
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
	done  chan struct{}
	tasks []Task
}

func (sup *Supervisor) Stop() (<-chan struct{}, error) {
	return sup.SubmitChan(func() {
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

		close(sup.done)
		// stop self
		<-sup.GenServer.Stop()
	})
}

func (sup *Supervisor) watch(task Task) {
	for {
		select {
		case <-sup.done:
			<-task.Stop()
			return
		case <-task.Ping():
		case <-time.After(100 * time.Millisecond):
			fmt.Println("task is stopped, will try to restart")
			<-task.Start()
		}
		// task is runing, wait for 100 milliseconds
		time.Sleep(100 * time.Millisecond)
	}
}

func (sup *Supervisor) SuperviseTasks() (<-chan struct{}, error) {
	return sup.SubmitChan(func() {
		for _, task := range sup.tasks {
			go sup.watch(task)
		}
	})
}

func (sup *Supervisor) StartTasks() (<-chan struct{}, error) {
	return sup.SubmitChan(func() {
		// start all tasks
		var wg sync.WaitGroup
		wg.Add(len(sup.tasks))
		for _, task := range sup.tasks {
			go func(t Task) {
				<-t.Start()
				wg.Done()
			}(task)
		}
		wg.Wait()
	})
}

func (sup *Supervisor) AddTasks(tasks ...Task) (<-chan struct{}, error) {
	return sup.SubmitChan(func() {
		sup.tasks = append(sup.tasks, tasks...)
	})
}

func TestSupervisor(t *testing.T) {
	kvm := &KVMServer{
		kv: make(map[string]string),
	}
	sup := new(Supervisor)
	sup.Start()
	sup.AddTasks(kvm)
	<-sup.StartTasks()
	<-sup.SuperviseTasks()

	kvm.Put("Hello", "100")
	fmt.Println(kvm.Get("Hello"))
	<-kvm.Stop()
	fmt.Println(kvm.Get("Hello"))
	<-sup.Stop()
}
