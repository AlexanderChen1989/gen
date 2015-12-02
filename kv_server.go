package main

type KVServer struct {
	evts chan func(KV) KV
}

func StartKVServer(init func() KV) *KVServer {
	kvs := &KVServer{
		evts: make(chan func(KV) KV),
	}

	go kvs.loop(init())

	return kvs
}

func (kvs *KVServer) loop(kv KV) {
	for evt := range kvs.evts {
		kv = evt(kv)
	}
}

func (kvs *KVServer) Put(k string, v string) {
	kvs.evts <- func(kv KV) KV {
		return KVStore.Put(kv, k, v)
	}
}

func (kvs *KVServer) Get(k string) string {
	ch := make(chan string, 1)
	kvs.evts <- func(kv KV) KV {
		v, kv := KVStore.Get(kv, k)
		ch <- v
		return kv
	}
	return <-ch
}
