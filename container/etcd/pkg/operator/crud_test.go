package operator_test

import (
	"context"
	"fmt"
	"orm/container/etcd/pkg/operator"

	"log"
	"math/rand"
	"strconv"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var prefix = "/test/crud/"

func TestTxn(t *testing.T) {
	ec := operator.NewEtcdClinet()
	txn := ec.Kv.Txn(context.TODO())
	key := prefix + "key" + strconv.Itoa(rand.Intn(100))
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "192.0.11.83")).Else(
		clientv3.OpGet(key),
	)
	resp, err := txn.Commit()
	if err != nil {
		log.Fatalln(err)
	}

	for _, res := range resp.Responses {
		for _, r := range res.GetResponseRange().Kvs {
			fmt.Println("key: ", string(r.Key), " value: ", string(r.Value))
		}
	}
}
func TestCrudPut(t *testing.T) {
	crud := operator.NewEtcdClinet()
	if crud == nil {
		t.Fatal()
	}
	for i := 0; i < 100; i++ {
		err := crud.Put(prefix+"key"+strconv.Itoa(i), "192.0.11.82")
		if err != nil {
			fmt.Println("key"+strconv.Itoa(i), err)
		}
	}
}

func TestCrudList(t *testing.T) {
	crud := operator.NewEtcdClinet()
	if crud == nil {
		t.Fatal()
	}
	list, err := crud.List(prefix)
	if err != nil {
		fmt.Println(err)
	}

	for _, item := range list {
		fmt.Printf("key:%v,value:%v", item.Key, item.Value)
	}
}

func TestDelete(t *testing.T) {
	crud := operator.NewEtcdClinet()
	if crud == nil {
		t.Fatal()
	}

	err := crud.Delete(prefix)
	if err != nil {
		fmt.Println(err)
	}
}
