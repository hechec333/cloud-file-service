package server_test

import (
	"context"
	"fmt"
	"log"
	"net"
	auth "orm/service/auth/rpc/types"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
)

type AuthGrpcServerImpl struct {
	auth.UnimplementedAuthServer
}

func NewAuthGrpcSever() *AuthGrpcServerImpl {
	return &AuthGrpcServerImpl{}
}

// func (au*AuthGrpcServerImpl)
func (au *AuthGrpcServerImpl) GetLoginAuth(ctx context.Context, req *auth.GetLoginAuthRequest) (*auth.GetLoginAuthResponce, error) {
	return &auth.GetLoginAuthResponce{
		AccessToken:  "xxxa",
		ExpireIn:     190539105,
		RefreshToken: "xxx1",
	}, nil
}
func (au *AuthGrpcServerImpl) CheckEmailAuth(ctx context.Context, req *auth.CheckEmailAuthRequest) (*auth.CheckEmailAuthResponce, error) {

	return &auth.CheckEmailAuthResponce{
		Success:   true,
		AuthToken: "xxxa",
	}, nil
}
func (au *AuthGrpcServerImpl) VertifyCaptcha(ctx context.Context, req *auth.VertifyCaptchaRequest) (*auth.VertifyCaptchaResponce, error) {

	return &auth.VertifyCaptchaResponce{
		Success: true,
	}, nil
}

func getNetListener(port int) *net.Listener {

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	fmt.Println(lis)
	return &lis
}

func StartAuthSever(port int) {

	g := grpc.NewServer()

	auth.RegisterAuthServer(g, NewAuthGrpcSever())

	if err := g.Serve(*getNetListener(port)); err != nil {
		panic(err)
	}
}

func GetAuthClient(port int) auth.AuthClient {

	addr := "localhost:" + strconv.Itoa(port)

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}

	return auth.NewAuthClient(conn)
}

// === RUN   TestAuthGrpc
// &{0xc000154300 {<nil> 0}}
// /home/hec/go/go-orm/service/auth/rpc/server/server_test.go:87: testing GetLoginAuth Call
// /home/hec/go/go-orm/service/auth/rpc/server/server_test.go:97: AccessToken:"xxxa" ExpireIn:190539105 RefreshToken:"xxx1"
// /home/hec/go/go-orm/service/auth/rpc/server/server_test.go:98: testing CheckEmailAuth Call
// /home/hec/go/go-orm/service/auth/rpc/server/server_test.go:108: Success:true AuthToken:"xxxa"
// /home/hec/go/go-orm/service/auth/rpc/server/server_test.go:115: testing VertifyCaptcha Call
// /home/hec/go/go-orm/service/auth/rpc/server/server_test.go:119: Success:true
// --- PASS: TestAuthGrpc (5.01s)
// PASS
// ok  	orm/service/auth/rpc/server	5.011s
func TestAuthGrpc(t *testing.T) {

	go StartAuthSever(8060)

	time.Sleep(5 * time.Second)

	cc := GetAuthClient(8060)

	t.Log("testing GetLoginAuth Call")

	res, err := cc.GetLoginAuth(context.Background(), &auth.GetLoginAuthRequest{
		UserId: 190801290,
	})

	if err != nil {
		t.Fail()
	}

	t.Log(res)
	t.Log("testing CheckEmailAuth Call")

	resx, err := cc.CheckEmailAuth(context.Background(), &auth.CheckEmailAuthRequest{
		Eid:  "ssxx",
		Code: "2341xwa",
	})

	if err != nil {
		t.Fail()
	}
	t.Log(resx)

	resf, err := cc.VertifyCaptcha(context.TODO(), &auth.VertifyCaptchaRequest{
		Cid:  "xr2",
		Code: "251x43",
	})

	t.Log("testing VertifyCaptcha Call")
	if err != nil {
		t.Fail()
	}
	t.Log(resf)
}
