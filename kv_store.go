package main

type KV map[string]string

func NewKVStore() KV {
	return make(KV)
}

type kvstore struct{}

func (_ kvstore) Get(kv KV, k string) (string, KV) {
	return kv[k], kv
}

func (_ kvstore) Put(kv KV, k, v string) KV {
	kv[k] = v
	return kv
}

var KVStore = kvstore{}
