package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	timeout   int = 2
	ttl       int = 10
	endPoints     = []string{"localhost:3379", "localhost:3380", "localhost:3381"}
)

type ServiceDiscovery struct {
	endPoints []string
	kv        clientv3.KV
	cli       *clientv3.Client
	ctx       context.Context
	mu        sync.Mutex
	refCount  int
}

var sd *ServiceDiscovery = nil

func GetServiceDiscovery() *ServiceDiscovery {
	if sd == nil {
		mu := sync.Mutex{}
		defer mu.Unlock()
		sd := &ServiceDiscovery{
			endPoints: endPoints,
			kv:        nil,
			cli:       nil,
			mu:        sync.Mutex{},
			refCount:  1,
		}

		err := sd.connect()

		if err != nil {
			return nil
		}

	}

	sd.mu.Lock()

	defer sd.mu.Unlock()

	sd.refCount++

	return sd
}

func (sd *ServiceDiscovery) Close() {

	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.refCount == 1 {
		sd.cli.Close()
	} else {
		sd.refCount--
	}
}

func (sd *ServiceDiscovery) connect() (err error) {
	sd.cli, err = clientv3.New(
		clientv3.Config{
			Endpoints:   sd.endPoints,
			DialTimeout: 5 * time.Second,
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	sd.kv = clientv3.NewKV(sd.cli)
	sd.ctx = context.Background()
	return
}

type EtcdServerRegister struct {
	svc      *ServiceDiscovery
	lease    clientv3.LeaseID
	leaseTTL int64
}

func NewEtcdServerRegiser() *EtcdServerRegister {
	et := &EtcdServerRegister{
		svc:      GetServiceDiscovery(),
		leaseTTL: int64(ttl),
	}

	resp, err := et.svc.cli.Grant(context.Background(), et.leaseTTL)

	if err != nil {
		panic(err)
	}

	et.lease = resp.ID

	return et
}

func (sd *EtcdServerRegister) Register(types string, target string, endpoint string) error {

	prefix := "rpc"
	key := prefix + "/" + types + "/" + target + "/" + endpoint
	_, err := sd.svc.cli.KV.Put(sd.svc.ctx, key, endpoint)
	if err != nil {
		return err
	}

	go sd.keepAlive()
	return nil
}

func (sd *EtcdServerRegister) keepAlive() {
	ticker := time.NewTicker(4 * time.Second)
	for {
		select {
		case <-sd.svc.ctx.Done():
			return
		case <-ticker.C:
			_, err := sd.svc.cli.KeepAliveOnce(sd.svc.ctx, sd.lease)
			if err != nil {
				panic(err)
			}
		}
	}
}
