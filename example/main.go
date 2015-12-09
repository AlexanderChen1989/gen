package main

import (
	"fmt"

	"github.com/AlexanderChen1989/gen"
	"golang.org/x/net/context"
)

type kvServer struct {
	gen.GenServer
	kv map[string]string
}

func (kv *kvServer) Put(k, v string) error {
	return kv.Submit(func(ctx context.Context) {
		kv.kv[k] = v
	})
}

func (kv *kvServer) Get(k string) (string, error) {
	ch := make(chan string, 1)
	err := kv.Submit(func(ctx context.Context) {
		ch <- kv.kv[k]
	})
	if err != nil {
		return "", err
	}
	return <-ch, nil
}

func main() {
	s := kvServer{
		GenServer: gen.New(nil),
		kv:        make(map[string]string),
	}
	s.Start()
	for i := 0; i < 100; i++ {
		k := fmt.Sprint(i)
		s.Put(k, k+"val")
	}
	for i := 0; i < 100; i++ {
		fmt.Println(s.Get(fmt.Sprint(i)))
	}
}
