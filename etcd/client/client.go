package client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

var (
	timeout int = 2
	ttl     int = 15
)

var Endpoints = []string{"localhost:3379", "localhost:3380", "localhost:3381"}

const (
	KEY_CREATE = int(0)
	KEY_UPDATE = int(1)
	KEY_DELETE = int(2)
	KEY_ERR    = int(3)
)

func SetTTL(ttlx int) {
	ttl = ttlx
}

type ServiceResolver struct {
	endPoints []string
	kv        clientv3.KV
	cli       *clientv3.Client
	ctx       context.Context
	lease     clientv3.Lease
	leaseTTL  int64
	mu        sync.Mutex
	refCount  int
}

var etcdresolver *ServiceResolver = nil

func GetServiceResolver() *ServiceResolver {

	if etcdresolver == nil {
		mu := sync.Mutex{}
		mu.Lock()
		defer mu.Unlock()
		sd := &ServiceResolver{
			endPoints: Endpoints,
			kv:        nil,
			cli:       nil,
			leaseTTL:  int64(ttl),
			mu:        sync.Mutex{},
			refCount:  1,
		}

		err := sd.connect()

		if err != nil {
			panic(err)
		}

		etcdresolver = sd
	}

	etcdresolver.mu.Lock()
	defer etcdresolver.mu.Unlock()

	etcdresolver.refCount++

	return etcdresolver
}

func (sd *ServiceResolver) Close() {

	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.refCount == 1 {
		sd.cli.Close()
	} else {
		sd.refCount--
	}

}

func (sd *ServiceResolver) connect() (err error) {
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
	sd.lease = clientv3.NewLease(sd.cli)
	return
}

func (sd *ServiceResolver) getSeviceEndpoints(keyprefix string) []resolver.Address {
	var addrList []resolver.Address

	resp, err := sd.kv.Get(sd.ctx, keyprefix, clientv3.WithPrefix())

	if err != nil {
		fmt.Println(err)
	} else {
		for i := range resp.Kvs {
			addrList = append(addrList, resolver.Address{Addr: strings.TrimPrefix(string(resp.Kvs[i].Key), keyprefix)})
		}
	}

	return addrList
}

const scheme = "etcd"

type GrpcResovlerBuilder struct {
}

func (gr *GrpcResovlerBuilder) Scheme() string {
	return scheme
}

func (grb *GrpcResovlerBuilder) Build(target resolver.Target,
	cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	cr := &GrpcResovler{
		svc:    GetServiceResolver(),
		cc:     cc,
		ctx:    context.Background(),
		target: &target,
	}
	go cr.watcher()
	return cr, nil
}

type GrpcResovler struct {
	svc    *ServiceResolver
	cc     resolver.ClientConn
	target *resolver.Target
	ctx    context.Context
}

func (gr *GrpcResovler) Close() {
	gr.ctx.Done()
	gr.svc.Close()
}

func (gr *GrpcResovler) watcher() {
	//初始化服务地址列表
	addrs := gr.svc.getSeviceEndpoints(gr.target.Endpoint())

	gr.cc.UpdateState(resolver.State{Addresses: addrs})
	//监听服务地址列表的变化
	rch := gr.svc.cli.Watch(context.Background(), gr.target.Endpoint(), clientv3.WithPrefix())
	for {
		select {
		case n := <-rch:
			for _, ev := range n.Events {
				addr := strings.TrimPrefix(string(ev.Kv.Key), gr.target.Endpoint())
				switch ev.Type {
				case mvccpb.PUT:
					if !exists(addrs, addr) {
						addrs = append(addrs, resolver.Address{Addr: addr})
						gr.cc.UpdateState(resolver.State{Addresses: addrs})
					}
				case mvccpb.DELETE:
					if s, ok := remove(addrs, addr); ok {
						addrs = s
						gr.cc.UpdateState(resolver.State{Addresses: addrs})
					}
				}
			}
		case <-gr.ctx.Done():
			return
		}
	}
}

func exists(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

func (gr *GrpcResovler) ResolveNow(resolver.ResolveNowOptions) {

	addrs := gr.svc.getSeviceEndpoints(gr.target.Endpoint())

	gr.cc.UpdateState(resolver.State{Addresses: addrs})
}
