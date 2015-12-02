package main

import "fmt"
import "./kv"

func main() {
	kvs := kv.NewKVServer()

	kvs.Put("Hello", "World")
	kvs.Put("Good", "Body")
	fmt.Println(kvs.Get("Hello"))
	fmt.Println(kvs.Get("Good"))

	<-kvs.Stop()
}
