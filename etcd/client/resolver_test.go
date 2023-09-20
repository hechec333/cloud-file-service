package client_test

import (
	"orm/etcd/client"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

func TestGrpcResolver(t *testing.T) {

	serviceName := "user"
	resolver.Register(client.GrpcResovlerBuilder{})
	conn, err := grpc.Dial("etcd:///rpc/" + serviceName)

}
