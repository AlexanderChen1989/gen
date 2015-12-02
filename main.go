package main

import "fmt"

func main() {
	kvs := StartKVServer(NewKVStore)
	kvs.Put("Hello", "World")
	kvs.Put("Good", "Body")
	fmt.Println(kvs.Get("Hello"))
	fmt.Println(kvs.Get("Good"))
}
