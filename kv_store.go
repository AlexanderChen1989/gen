package main

type KV map[string]string

func NewKVStore() KV {
	return make(KV)
}

type kvstore struct{}

func (_ kvstore) Get(kv KV, k string) string {
	return kv[k]
}

func (_ kvstore) Put(kv KV, k, v string) {
	kv[k] = v
}

var KVStore = kvstore{}
