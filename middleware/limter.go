package middleware

import (
	"fmt"
	"orm/dao/cache"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// type LeakyBucket struct {
// 	rate       float64 //固定每秒出水速率
// 	capacity   float64 //桶的容量
// 	water      float64 //桶中当前水量
// 	lastLeakMs int64   //桶上次漏水时间戳 ms

// 	lock sync.Mutex
// }

// func (l *LeakyBucket) Allow() bool {
// 	l.lock.Lock()
// 	defer l.lock.Unlock()

// 	now := time.Now().UnixNano() / 1e6
// 	eclipse := float64((now - l.lastLeakMs)) * l.rate / 1000 //先执行漏水
// 	l.water = l.water - eclipse                              //计算剩余水量
// 	l.water = math.Max(0, l.water)                           //桶干了
// 	l.lastLeakMs = now
// 	if (l.water + 1) < l.capacity {
// 		// 尝试加水,并且水还未满
// 		l.water++
// 		return true
// 	} else {
// 		// 水满，拒绝加水
// 		return false
// 	}
// }

// func (l *LeakyBucket) Set(r, c float64) {
// 	l.rate = r
// 	l.capacity = c
// 	l.water = 0
// 	l.lastLeakMs = time.Now().UnixNano() / 1e6
// }

const MaxWait = 10

type RateLimter struct {
	rate int
	cap  int
	key  string
	wait chan struct{}
}

var registerMaps = make(map[string]*RateLimter)

func GenRaterLimiter(name string, rate int, upper int) *RateLimter {
	rr := RateLimter{
		rate: rate,
		key:  name,
		cap:  upper,
		wait: make(chan struct{}, rate),
	}

	registerMaps[name] = &rr
	return &rr
}

func LimiterInit() {
	t := time.NewTicker(50 * time.Millisecond)
	if len(registerMaps) != 0 {
		fmt.Println("starting rateLimit goroutine...")
		go func() {
			<-t.C
			for _, v := range registerMaps {
				sf, _ := cache.GetKey("lm:" + v.key)
				cap, _ := strconv.Atoi(sf)
				leak := v.rate * 50 / 1000
				if leak < 1 {
					leak = 1
				}
				if cap >= v.cap {
					for i := 0; i < leak; i++ {
						v.wait <- struct{}{}
					}
				}
				cap = cap - leak
				if cap < 0 {
					cap = 0
				}

				cache.SetKey("lm:"+v.key, strconv.Itoa(cap), 0)
			}
		}()
	}

}

func CaptchaRateLimte(ctx *gin.Context) {

	v := registerMaps["captcha"]
	sf, _ := cache.GetKey("lm:captcha")

	cap, _ := strconv.Atoi(sf)

	if cap >= v.cap {

		select {
		case <-time.After(10 * time.Second):
			ctx.Abort()
			ctx.HTML(400, "too many request", gin.H{
				"title": "slow responce",
				"msg":   "",
			})
		case <-v.wait:
		}
	}

	ctx.Next()
}
