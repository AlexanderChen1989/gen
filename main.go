package main

import (
	"fmt"

	"github.com/AlexanderChen1989/gen/kv"
)

func main() {
	kvs := kv.NewKVServer()

	kvs.Put("Hello", "World")
	kvs.Put("Good", "Body")
	fmt.Println(kvs.Get("Hello"))
	fmt.Println(kvs.Get("Good"))

	<-kvs.Stop()
}
