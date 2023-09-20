package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var (
	deadline time.Time
	vars     string
	dead     int
	mu       sync.Mutex
)

var logger *log.Logger
var xmu sync.Mutex

func INFO(format string, args ...any) {
	if logger == nil {
		xmu.Lock()
		defer xmu.Unlock()

		if logger == nil {
			f, _ := os.OpenFile("/var/log/go/main.log", os.O_CREATE, 0666)
			logger = log.New(f, "main.go ", log.LstdFlags)
		}
	}

	logger.Printf("[INFO] "+format, args...)
}
func FAIL(format string, args ...any) {
	if logger == nil {
		xmu.Lock()
		defer xmu.Unlock()

		if logger == nil {
			f, _ := os.OpenFile("/var/log/go/main.log", os.O_CREATE, 0666)
			logger = log.New(f, "main.go ", log.LstdFlags)
		}
	}

	logger.Printf("[FAIL] "+format, args...)
}

func setDeadline() {
	adead := os.Getenv("GO_DEEMON_DEADLINE")
	if adead == "" {
		deadline = time.Now().Add(5 * time.Second)
	} else {
		dead, _ = strconv.Atoi(adead)
		deadline = time.Now().Add(time.Duration(dead * int(time.Second)))
	}
}
func can() {
	mu.Lock()
	defer mu.Unlock()
	deadline = time.Now().Add(time.Duration(dead * int(time.Second)))
}
func main() {
	setDeadline()
	vars = os.Getenv("GO_DAEMON_SOCK")
	if vars == "" {
		vars = "/tmp/go-deamon.sock"
	}
	l, err := net.Listen("unix", vars)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.TODO())

	// 清理sock文件。
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
		case <-ctx.Done():
		}
		os.Remove(vars)
		os.Exit(1)
	}()

	// 接受传入的连接。
	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}

	// 监听worker的活动
	go func(conn net.Conn) {
		defer conn.Close()
		for {
			// 为传入的数据创建一个缓冲区。
			buf := make([]byte, 4096)

			// 从连接中读取数据。
			_, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			if string(buf) == "0" {
				cancel()
				INFO("receive worker quit signals")
				break
			}
			can()
		}
	}(conn)
	// 重启worker
	go func() {
		tick := time.Tick(time.Second)

		for {
			select {
			case <-tick:
				mu.Lock()
				if time.Now().After(deadline) {
					cmd := exec.Command("sh", "-c", "/go-nginx/main.go")
					err = cmd.Run()
					if err != nil {
						FAIL("", err)
						os.Exit(1)
					}
				}
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	time.Sleep(3 * time.Second)

	INFO("daemon quit")
}
