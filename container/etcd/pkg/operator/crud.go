package operator

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	timeout int = 1
	address     = []string{"0.0.0.0:3379", "0.0.0.0:3380", "0.0.0.0:3381"}
)

type EtcdClient struct {
	Cli       *clientv3.Client
	endPoints []string
	Kv        clientv3.KV
	timeout   int64
}
type GetMeta struct {
	Key        string
	Value      string
	ModVersion int64
	Version    int64
	Lease      int64
}

func NewEtcdClinet() *EtcdClient {
	cli := &EtcdClient{
		Cli:       nil,
		endPoints: address,
		timeout:   int64(timeout),
	}

	err := cli.connect()
	if err != nil {
		return nil
	}
	return cli
}

func (c *EtcdClient) connect() (err error) {
	c.Cli, err = clientv3.New(
		clientv3.Config{
			Endpoints:   c.endPoints,
			DialTimeout: 5 * time.Second,
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Kv = clientv3.NewKV(c.Cli)
	return
}

func (c *EtcdClient) Close() error {
	return c.Cli.Close()
}
func (c *EtcdClient) Get(key string) (string, error) {
	ctx, h := context.WithTimeout(context.TODO(), 3*time.Second)
	resp, err := c.Kv.Get(ctx, key, clientv3.WithPrefix())
	h()
	if err != nil {
		return "", err
	}
	return string(resp.Kvs[0].Value), nil
}
func (c *EtcdClient) List(prefix string) ([]GetMeta, error) {
	ctx, h := context.WithTimeout(context.TODO(), 3*time.Second)
	resp, err := c.Kv.Get(ctx, prefix, clientv3.WithPrefix())
	h()
	res := []GetMeta{}
	if err != nil {
		return res, err
	}

	for _, v := range resp.Kvs {
		res = append(res, GetMeta{
			Key:        string(v.Key),
			Value:      string(v.Value),
			ModVersion: v.ModRevision,
			Version:    v.Version,
			Lease:      v.Lease,
		})
	}
	return res, nil
}
func (c *EtcdClient) Put(key string, v interface{}) error {
	ctx, h := context.WithTimeout(context.TODO(), time.Duration(timeout*int(time.Second)))
	_, err := c.Kv.Put(ctx, key, fmt.Sprintf("%s", v))
	h()
	return err
}

func (c *EtcdClient) Delete(key string) error {
	ctx, h := context.WithTimeout(context.TODO(), 3*time.Second)
	_, err := c.Kv.Delete(ctx, key, clientv3.WithPrefix())
	h()
	return err
}
