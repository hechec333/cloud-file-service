package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Config struct {
	Workers     int `xml:"workers"`
	Port        int `xml:"listen"`
	Connections int `xml:"connenctions"`
	Timeout     int `xml:"keepalive"`
}
type Targets struct {
	Name    string `xml:"name"`
	Address string `xml:"address"`
}
type Service struct {
	Name    string `xml:"name"`
	Address string `xml:"address"`
	Trim    bool   `xml:"trim-url"`
}
type Nginx struct {
	Services []Service `xml:"services"`
	Targets  Targets   `xml:"targets"`
	Config   Config    `xml:"Config"`
}

var nginx *Nginx

func newNginxConfig(r io.Reader) *Nginx {
	d := xml.NewDecoder(r)

	n := &Nginx{}

	d.Decode(n)

	return n
}

func GetNginxConf() *Nginx {
	if nginx == nil {
		prefix := os.Getenv("NGINX_XML")
		if prefix == "" {
			prefix = "/etc/nginx/"
		}
		f, err := os.Open(prefix + "nginx.xml")
		if err != nil {
			panic(err)
		}
		nginx = newNginxConfig(f)
	}

	return nginx
}

func isFileExist(pt string) bool {
	if _, err := os.Stat(pt); err != nil {
		return os.IsNotExist(err)
	}
	return true
}

func GenNginxConf(out string) error {
	path := out + "/nginx.conf"
	tmpl, err := template.New("./nginx.conf.tmpl").Funcs(
		template.FuncMap{
			"getAddr": func(addr string) []string {
				return strings.Split(addr, ";")
			},
		},
	).ParseFiles("./nginx.conf.tmpl")
	if err != nil {
		FAIL("", err)
		return err
	}
	cfg := GetNginxConf()
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, *cfg)

	if err != nil {
		FAIL("", err)
		return err
	}

	if isFileExist(path) {
		os.Remove(path)
	}

	file, _ := os.Create(path)

	file.WriteString(buf.String())
	return nil
}

func ReloadNginx() error {
	err := GenNginxConf("/etc/nginx")
	if err != nil {
		return err
	}

	cmd := exec.Command("nginx", "-s", "reload")

	return cmd.Run()
}

func LoadNginx() error {
	err := GenNginxConf("/etc/nginx")
	if err != nil {
		return err
	}
	cmd := exec.Command("nginx", "-c", "/etc/nginx/nginx.conf", "-g", "daemon off;")

	return cmd.Run()
}
func connect(cfg *Nginx) (*clientv3.Client, error) {
	endpoints := strings.Split(cfg.Targets.Address, ";")

	return clientv3.New(
		clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
		},
	)
}
func EtcdRoutine(ctx context.Context) {
	prefix := "/service"
	INFO("etcd watch routine start,watching all %v changes", prefix)
	cfg := GetNginxConf()
	cli, err := connect(cfg)
	if err != nil {
		panic(err)
	}
	apply := make(chan struct{})
	for _, v := range cfg.Services {
		go watch(ctx, cli, v.Name, prefix, apply)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-apply:
				var err error
				if isNginxAlive() {
					err = ReloadNginx()
				} else {
					err = LoadNginx()
				}
				INFO("apply new config file error %v", err)
			}
		}
	}()
}

func watch(ctx context.Context, cli *clientv3.Client, service, prefix string, no chan struct{}) {
	long := cli.Watch(ctx, prefix+"/"+service, clientv3.WithPrefix())
	for r := range long {
		for _, v := range GetNginxConf().Services {
			if v.Name == service {
				INFO("discover new sevice entry of %v", service)
				addrs := strings.Split(v.Address, ";")
				for _, item := range r.Events {
					if item.IsCreate() {
						addrs = append(addrs, path.Base(string(item.Kv.Key))+":"+string(item.Kv.Value)) // etcd:2379
					} else if item.IsModify() {
						for i, vv := range addrs {
							if vv == path.Base(string(item.Kv.Key))+":"+string(item.PrevKv.Value) {
								addrs[i] = string(item.Kv.Key) + ":" + string(item.Kv.Value)
								break
							}
						}
					} else {
						for i, vv := range addrs {
							if vv == path.Base(string(item.Kv.Key))+":"+string(item.PrevKv.Value) {
								addrs = append(addrs[:i], addrs[i+1:]...)
								break
							}
						}
					}
				}

				v.Address = strings.Join(addrs, ";")

				no <- struct{}{}
			}
		}
	}

}
func isNginxAlive() bool {
	cmd := exec.Command("sh", "-c", "netstat -anp | grep :80")

	out, err := cmd.Output()

	if err != nil {
		return false
	}

	output := strings.TrimSpace(string(out))
	return strings.Contains(output, "nginx")
}
func StartDaemon(ctx context.Context) {
	sock := "/tmp/go-daemon.sock"

	if isFileExist(sock) {
		INFO("Daemon already start sock in %v", sock)
		return
	}
	env := os.Getenv("GROOT")
	cmd := exec.Command("go", "run", fmt.Sprintf("%v/daemon/daemon.go", env))

	cmd.Env = append(cmd.Env, "GO_DEEMON_DEADLINE=10", "GO_DAEMON_SOCK="+sock)

	b, err := cmd.Output()
	if err != nil {
		FAIL("", err)
		return
	}

	INFO("daemon resp %v", string(b))

	go func() {
		conn, err := net.Dial("unix", sock)
		if err != nil {
			FAIL("", err)
			return
		}
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				conn.Write([]byte("1"))
			case <-ctx.Done():
				conn.Write([]byte("0"))
				time.Sleep(time.Second)
				return
			}
		}
	}()
}
func waitSignal() {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
}

var logger *log.Logger
var mu sync.Mutex

func INFO(format string, args ...any) {
	if logger == nil {
		mu.Lock()
		defer mu.Unlock()

		if logger == nil {
			f, _ := os.OpenFile("/var/log/go/main.log", os.O_CREATE, 0666)
			logger = log.New(f, "main.go ", log.LstdFlags)
		}
	}

	logger.Printf("[INFO] "+format, args...)
}
func FAIL(format string, args ...any) {
	if logger == nil {
		mu.Lock()
		defer mu.Unlock()

		if logger == nil {
			f, _ := os.OpenFile("/var/log/go/main.log", os.O_CREATE, 0666)
			logger = log.New(f, "main.go ", log.LstdFlags)
		}
	}

	logger.Printf("[FAIL] "+format, args...)
}
func main() {
	var (
		err    error
		daemon string
	)
	flag.StringVar(&daemon, "d", "off", "")

	flag.Parse()
	if isNginxAlive() {
		INFO("Nginx is already running!")
		err = ReloadNginx()
	} else {
		err = LoadNginx()
	}

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	if daemon == "on" {
		INFO("daemon flag [on],start daemon")
		StartDaemon(ctx)
	}
	EtcdRoutine(ctx)

	waitSignal()

	cancel()

	time.Sleep(3 * time.Second)
}
