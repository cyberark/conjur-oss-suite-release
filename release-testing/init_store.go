package main

import "test/kv"

var storeClient *kv.StoreClient

func init() {
	var err error
	storeClient, err = kv.DefaultStoreClient()
	if err != nil {
		panic(err)
	}
}
