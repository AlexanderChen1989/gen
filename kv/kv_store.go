package kv

type kvStore map[string]string

func newKVStore() kvStore {
	return make(kvStore)
}

func kvStoreGet(kv kvStore, k string) (string, kvStore) {
	return kv[k], kv
}

func kvStorePut(kv kvStore, k, v string) kvStore {
	kv[k] = v
	return kv
}
