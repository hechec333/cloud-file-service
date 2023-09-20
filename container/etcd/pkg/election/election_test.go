package election_test

import (
	"context"
	"fmt"
	"orm/container/etcd/pkg/election"

	"math/rand"
	"strconv"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestElection(t *testing.T) {

	nums := 3
	cfg := clientv3.Config{
		Endpoints: []string{"0.0.0.0:3379", "0.0.0.0:3380", "0.0.0.0:3381"},
	}
	for i := 0; i < nums; i++ {
		ccc, err := election.NewElectionControllor(cfg, "node"+strconv.Itoa(i))
		if err != nil {
			fmt.Println(i, ":", err)
			continue
		}
		go func(cc *election.ElectionControllor) {
			time.Sleep(time.Duration(rand.Intn(5) * int(time.Second)))
			ch := make(chan election.ElecResult, 1)
			go cc.Election(context.TODO(), ch)
			for s := range ch {
				if s.Error() {
					fmt.Println(s.Err)
					return
				} else if s.FirstQuroam() {
					fmt.Println(cc.SelfMeta, ": ", cc.Charactor())
				} else if s.ToMaster() {
					fmt.Println(cc.SelfMeta, ": ", cc.Charactor())
				}
			}
		}(ccc)
	}
	time.Sleep(time.Hour)
}
