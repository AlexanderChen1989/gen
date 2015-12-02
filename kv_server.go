package main

type KVServer struct {
	kv   KV
	evts chan func()
}

func StartKVServer(init func() KV) *KVServer {
	kvs := &KVServer{
		evts: make(chan func()),
	}
	kvs.kv = init()

	go kvs.loop()

	return kvs
}

func (kvs *KVServer) Put(k, v string) {
	kvs.put(kvs.kv, k, v)
}

func (kvs *KVServer) Get(k string) string {
	return kvs.get(kvs.kv, k)
}

func (kvs *KVServer) loop() {
	for evt := range kvs.evts {
		evt()
	}
}

func (kvs *KVServer) put(kv KV, k string, v string) {
	kvs.evts <- func() {
		kv[k] = v
	}
}

func (kvs *KVServer) get(kv KV, k string) string {
	ch := make(chan string, 1)
	kvs.evts <- func() {
		ch <- kv[k]
	}
	return <-ch
}
