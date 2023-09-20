package election

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	TO_LEADER   int = 1
	TO_FOLLOWER int = 2
)

var mode = []string{"Leader", "Follwer", "None"}
var prefix = "/test_election"

type ElecResult struct {
	ElecIndex    int
	ChangeEvents int
	PreState     string
	CurrentState string
	Err          error
}

func (er *ElecResult) Error() bool {
	return er.Err != nil
}
func (er *ElecResult) FirstQuroam() bool {
	return er.ElecIndex == 1
}
func (er *ElecResult) ToMaster() bool {
	return er.ChangeEvents == TO_LEADER
}
func (er *ElecResult) ToFollower() bool {
	return er.ChangeEvents == TO_FOLLOWER
}

type ElectionControllor struct {
	SelfMeta   string
	LeaderMeta string
	Conf       *clientv3.Config
	cli        *clientv3.Client
	lease      clientv3.Lease
	leaseId    clientv3.LeaseID
	txn        clientv3.Txn
	cancel     context.CancelFunc
	state      string
	index      int
}

func NewElectionControllor(cfg clientv3.Config, meta string) (*ElectionControllor, error) {
	ec := &ElectionControllor{
		Conf:     &cfg,
		state:    mode[2],
		index:    0,
		SelfMeta: meta,
	}
	return ec, ec.init()
}
func (ec *ElectionControllor) Charactor() string {
	return ec.state
}
func (ec *ElectionControllor) init() error {
	var err error
	ec.cli, err = clientv3.New(*ec.Conf)
	if err != nil {
		return err
	}
	ec.lease = clientv3.NewLease(ec.cli)

	leaseResp, err := ec.lease.Grant(context.TODO(), 20)
	if err != nil {
		return err
	}
	ec.leaseId = leaseResp.ID
	var ctx context.Context
	ctx, ec.cancel = context.WithCancel(context.TODO())
	_, err = ec.lease.KeepAlive(ctx, ec.leaseId)
	return err
}

func (ec *ElectionControllor) Election(cx context.Context, elec chan ElecResult) {
	err := ec.quroam(elec)
	if err != nil {
		elec <- ElecResult{
			Err: err,
		}
		return
	}
	ch := make(chan int, 1)
	ctx, h := context.WithCancel(context.TODO())
	go ec.obeserve(ctx, ch)
	for {
		select {
		case <-cx.Done():
			h()
			elec <- ElecResult{
				ElecIndex: ec.index,
				Err:       cx.Err(),
			}
			return
		case r := <-ch:
			// 主节点故障
			if r == 1 {
				ec.quroam(elec)
				if ec.isActive() {
					h()
				}
			}
		}
	}
}

func (ec *ElectionControllor) isActive() bool {
	return ec.state == mode[0]
}
func (ec *ElectionControllor) quroam(elec chan ElecResult) error {
	ec.index++
	ec.txn = clientv3.NewKV(ec.cli).Txn(context.TODO())
	ec.txn.If(
		clientv3.Compare(clientv3.CreateRevision(prefix), "=", 0),
	).Then(
		clientv3.OpPut(prefix, ec.SelfMeta, clientv3.WithLease(ec.leaseId)),
	).Else(
		clientv3.OpGet(prefix),
	)

	resp, err := ec.txn.Commit()
	if err != nil {
		return err
	}
	if resp.Succeeded {
		if ec.state == mode[1] {
			elec <- ElecResult{
				ElecIndex:    ec.index,
				ChangeEvents: TO_LEADER,
				PreState:     ec.state,
				CurrentState: mode[0],
			}
		}
		ec.state = mode[0]
		ec.LeaderMeta = ec.SelfMeta
	} else {
		if ec.isActive() || ec.index == 1 {
			elec <- ElecResult{
				ElecIndex:    ec.index,
				ChangeEvents: TO_FOLLOWER,
				PreState:     ec.state,
				CurrentState: mode[1],
			}
		}
		ec.state = mode[1]
		ec.LeaderMeta = string(resp.Responses[0].GetResponseRange().Kvs[0].Value)
	}
	return nil
}
func (ec *ElectionControllor) obeserve(ctx context.Context, resign chan int) {
	chanw := ec.cli.Watch(ctx, prefix)
	for {
		select {
		case <-ctx.Done():
			return
		case w := <-chanw:
			for _, r := range w.Events {
				if r.Type == mvccpb.DELETE {
					resign <- 1
				} else if r.IsCreate() {
					resign <- 0
				}
			}
		}
	}
}

func (ec *ElectionControllor) Close() {
	ec.cancel()
	ec.lease.Revoke(context.Background(), ec.leaseId)
	fmt.Println("etcd election format")
}
