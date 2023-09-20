package dlock_test

import (
	"fmt"
	"orm/container/etcd/pkg/dlock"

	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var endPoints = []string{"0.0.0.0:3379", "0.0.0.0:3380", "0.0.0.0:3381"}

func Testdlock(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   endPoints,
		DialTimeout: 5 * time.Second,
	}
	mu1 := dlock.EtcdMutex{
		Conf: conf,
		Ttl:  10,
		Key:  "/etcd_lock",
	}
	mu2 := dlock.EtcdMutex{
		Conf: conf,
		Ttl:  20,
		Key:  "/etcd_lock",
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		time.Sleep(time.Duration((rand.Intn(5) + 1) * int(time.Second)))
		for {
			if y, err := mu1.Lock(); err != nil {
				if y {
					log.Println("mu1 locking")
					time.Sleep(5 * time.Second)
					log.Println("mu1 unlock")
					mu1.UnLock()
					wg.Done()
					break
				}
			}
		}

	}()

	go func() {
		time.Sleep(time.Duration((rand.Intn(5) + 1) * int(time.Second)))
		for {
			if y, err := mu2.Lock(); err != nil {
				if y {
					log.Println("mu2 locking")
					time.Sleep(5 * time.Second)
					mu1.UnLock()
					fmt.Println("mu2 unlock")
					wg.Done()
					break
				}
			}
		}

	}()
	wg.Wait()
}
