package discovery_test

import (
	"fmt"
	"orm/container/etcd/pkg/discovery"
	"testing"
)

var endPoints = []string{"0.0.0.0:3379", "0.0.0.0:3380", "0.0.0.0:3381"}

func TestRegister(t *testing.T) {
	sd := discovery.NewServiceDiscovery(endPoints)
	m1, err := sd.RegisterService("nginx/v1@node1", "182.92.114.28")
	if err != nil {
		fmt.Println(err)
		t.Fatal()
	}
	fmt.Println(m1.Key(), " success ")
	m2, err := sd.RegisterService("nginx/v1@node2", "182.92.114.29")

	if err != nil {
		fmt.Println(err)
		t.Fatal()
	}
	fmt.Println(m2.Key(), " success ")
}

func TestGet(t *testing.T) {
	d := discovery.NewServiceDiscovery(endPoints)
	arr, err := d.GetServiceEntries(discovery.ServicePrefix + "nginx")
	if err != nil {
		fmt.Println(err)
		t.Fatal()
	}
	for _, v := range arr {
		fmt.Println(v.Prefix, ":", v.ServiceName, " ", v.ServiceIp)
	}
}

func TestWatchKeys(t *testing.T) {
	d := discovery.NewServiceDiscovery(endPoints)
	d.WatchRange(discovery.ServicePrefix+"nginx", func(wm discovery.WatchMeta) {
		fmt.Println(wm.Code, "  ", string(wm.Data.Key), " ", string(wm.Data.Value))
	})
}
