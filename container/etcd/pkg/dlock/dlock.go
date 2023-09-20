package dlock

import (
	"context"

	"go.etcd.io/etcd/api/v3/mvccpb"
	v3 "go.etcd.io/etcd/client/v3"
)

/* 1.利用租约在etcd集群中创建一个key，这个key有两种形态，存在和不存在，而这两种形态就是互斥量
 * 2.如果这个key不存在，那么线程创建key，成功则获取到锁，该key就为存在状态。
 * 3.如果该key已经存在，那么线程就不能创建key，则获取锁失败
 *
 */

type EtcdMutex struct {
	Conf    v3.Config
	Ttl     int64
	Key     string
	cli     *v3.Client
	cancel  context.CancelFunc
	lease   v3.Lease
	leaseID v3.LeaseID
	txn     v3.Txn
}

func (em *EtcdMutex) init() error {
	var err error
	cc, err := v3.New(em.Conf)
	if err != nil {
		return err
	}
	em.cli = cc
	em.txn = v3.NewKV(cc).Txn(context.TODO())
	em.lease = v3.NewLease(cc)
	resp, err := em.lease.Grant(context.TODO(), em.Ttl)
	if err != nil {
		return err
	}
	var ctx context.Context
	ctx, em.cancel = context.WithCancel(context.TODO())
	em.leaseID = resp.ID
	_, err = em.lease.KeepAlive(ctx, em.leaseID)
	return err
}

// 尝试获取锁，不保证成功
func (em *EtcdMutex) Lock() (bool, error) {
	err := em.init()
	if err != nil {
		return false, err
	}
	//原地比较键是否被创建，如果没有被创建，则创建
	em.txn.If(
		v3.Compare(v3.CreateRevision(em.Key), "=", 0),
	).Then(
		v3.OpPut(em.Key, "", v3.WithLease(em.leaseID)),
	).Else()
	tnresp, err := em.txn.Commit()
	if err != nil {
		return false, err
	}
	return tnresp.Succeeded, nil
}

// 释放锁
func (em *EtcdMutex) UnLock() {
	em.cancel()
	em.lease.Revoke(context.TODO(), em.leaseID)
}

// 阻塞直到获取锁
func (em *EtcdMutex) SyncLock(ctx context.Context) bool {
	chanc := em.cli.Watch(ctx, em.Key)
	for {
		select {
		case r := <-chanc:
			for _, rr := range r.Events {
				if rr.Type == mvccpb.DELETE {
					yes, err := em.Lock()
					if err != nil {
						return false
					} else if yes {
						goto r
					}
				}
			}
		case <-ctx.Done():
			return false
		}
	}
r:
	return true
}
