package discovery

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	timeout int = 2
	ttl     int = 15
)
var logger *log.Logger = nil

const (
	KEY_CREATE = int(0)
	KEY_UPDATE = int(1)
	KEY_DELETE = int(2)
	KEY_ERR    = int(3)
)

func init() {
	fmt.Println("initing logger in ", getPkgName())
	logger = log.Default()
}

type pkgs struct {
}

func getPkgName() string {
	return reflect.TypeOf(pkgs{}).PkgPath()
}
func SetLogger(ll *log.Logger) {
	logger = ll
}
func SetTimeOut(t int) {
	timeout = t
}

type WatchCode int

func (c WatchCode) String() string {
	switch c {
	case WatchCode(KEY_CREATE):
		return "KEY_CREATE"
	case WatchCode(KEY_UPDATE):
		return "KEY_UPDATE"
	case WatchCode(KEY_DELETE):
		return "KEY_DELETE"
	case WatchCode(KEY_ERR):
		return "KEY_ERR"
	default:
		goto r
	}
r:
	return ""
}

type WatchMeta struct {
	Code WatchCode
	Data *mvccpb.KeyValue
}

type ServiceDiscovery struct {
	endPoints []string
	kv        clientv3.KV
	cli       *clientv3.Client
	ctx       context.Context
	lease     clientv3.Lease
	leaseTTL  int64
}
type ServiceMeta struct {
	sd          *ServiceDiscovery
	Prefix      string
	ServiceName string
	LeaseID     clientv3.LeaseID
	ClusterID   uint64
}
type ServiceEntry struct {
	Prefix      string
	ServiceName string
	ServiceIp   string
}

func NewServiceDiscovery(addr []string) *ServiceDiscovery {
	sd := &ServiceDiscovery{
		endPoints: addr,
		kv:        nil,
		cli:       nil,
		leaseTTL:  int64(ttl),
	}

	err := sd.connect()

	if err != nil {
		logger.Println(err)
		return nil
	}
	return sd
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
	sd.lease = clientv3.NewLease(sd.cli)
	return
}

// 一个ServiceDiscovery 只能注册一次
func (sd *ServiceDiscovery) RegisterService(Sname, value string) (*ServiceMeta, error) {
	// sd.lease = clientv3.NewLease(sd.cli)
	outline, h := context.WithTimeout(sd.ctx, time.Duration(timeout*int(time.Second)))
	leaseResp, err := sd.lease.Grant(outline, sd.leaseTTL)
	h()
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	meta := &ServiceMeta{
		Prefix:      ServicePrefix,
		ServiceName: Sname,
		LeaseID:     leaseResp.ID,
	}
	hs, err := sd.kv.Put(sd.ctx, ServicePrefix+Sname, value)
	if err != nil {
		return nil, err
	}
	meta.ClusterID = hs.Header.ClusterId
	meta.sd = sd

	go func() {
		defer func() {
			w := recover()
			logger.Println(w)
		}()
		meta.leaseKeepAlive()
	}()
	return meta, nil
}

func (sd *ServiceDiscovery) GetServiceEntries(prefix string) ([]*ServiceEntry, error) {
	cc, h := context.WithTimeout(sd.ctx, time.Duration(timeout*int(time.Second)))
	vv, err := sd.kv.Get(cc, prefix, clientv3.WithPrefix())
	h()
	res := []*ServiceEntry{}
	if err != nil {
		return res, err
	}

	for _, v := range vv.Kvs {
		k := string(v.Key)
		kk := strings.Replace(k, prefix, "", -1)
		if _, b := Ipvalidator(v.Value); !b {
			logger.Println(prefix+kk, " has unformat ip address:", string(v.Value))
			continue
		}
		res = append(res, &ServiceEntry{
			Prefix:      prefix,
			ServiceName: kk,
			ServiceIp:   string(v.Value),
		})
	}
	return res, nil
}

func ResolveWatchState(e *clientv3.Event) int {

	if e.IsCreate() {
		return KEY_CREATE
	} else if e.IsModify() {
		return KEY_UPDATE
	} else if e.Type == mvccpb.DELETE {
		return KEY_DELETE
	} else {
		return KEY_ERR
	}
}

type WatchCall func(WatchMeta)

func (sd *ServiceDiscovery) Watch(key string, invoker WatchCall) {
	sh := sd.cli.Watch(context.Background(), key)
	for r := range sh {
		for _, item := range r.Events {
			logger.Println("key:", string(item.Kv.Key), " / ", item.Type.String())
			invoker(WatchMeta{
				Code: WatchCode(ResolveWatchState(item)),
				Data: item.Kv,
			})
		}
	}
}
func (sd *ServiceDiscovery) WatchRange(prefix string, H WatchCall) {
	sh := sd.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for r := range sh {
		for _, item := range r.Events {
			logger.Println("key:", string(item.Kv.Key), " / ", item.Type.String())
			H(WatchMeta{
				Code: WatchCode(ResolveWatchState(item)),
				Data: item.Kv,
			})
		}
	}
}
func (sm *ServiceMeta) leaseKeepAlive() {
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-tick.C:
			_, err := sm.sd.lease.KeepAliveOnce(context.Background(), sm.LeaseID)
			if err != nil {
				panic(err)
			}
		}
	}
}
func (sm *ServiceMeta) Key() string {
	return sm.Prefix + sm.ServiceName
}
