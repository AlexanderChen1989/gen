package kv

type KVServer struct {
	evts chan func(kvStore) kvStore
}

func NewKVServer() *KVServer {
	return startKVServer(newKVStore)
}

func startKVServer(init func() kvStore) *KVServer {
	kvs := &KVServer{
		evts: make(chan func(kvStore) kvStore),
	}

	go kvs.loop(init())

	return kvs
}

func (kvs *KVServer) Ping() <-chan struct{} {
	ch := make(chan struct{})
	kvs.evts <- func(kv kvStore) kvStore {
		close(ch)
		return kv
	}
	return ch
}

func (kvs *KVServer) Stop() <-chan struct{} {
	ch := make(chan struct{})
	kvs.evts <- func(kv kvStore) kvStore {
		close(kvs.evts)
		close(ch)
		return kv
	}
	return ch
}

func (kvs *KVServer) Put(k string, v string) {
	kvs.evts <- func(kv kvStore) kvStore {
		return kvStorePut(kv, k, v)
	}
}

func (kvs *KVServer) Get(k string) string {
	ch := make(chan string, 1)
	kvs.evts <- func(kv kvStore) kvStore {
		v, kv := kvStoreGet(kv, k)
		ch <- v
		return kv
	}
	return <-ch
}

func (kvs *KVServer) loop(kv kvStore) {
	for evt := range kvs.evts {
		kv = evt(kv)
	}
}
