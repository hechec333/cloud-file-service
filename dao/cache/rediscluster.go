package cache

import (
	"bufio"
	"fmt"
	"io"
	"orm/config"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var RedisCluster *redis.ClusterClient

func readClusterInfo() []string {
	path := "./container/var/.redis"
	file, _ := os.Open(path)
	strs := []string{}
	rf := bufio.NewReader(file)

	for {
		line, err := rf.ReadString('\n')

		if err == io.EOF {
			break
		}

		if strings.Split(line, " ")[1] == "master\n" {
			strs = append(strs, strings.Split(line, " ")[0])
		}
		strs = append(strs, strings.Split(line, " ")[0])
	}
	fmt.Println(strs)
	return strs
}

// 1.更改redis镜像内部支持dns解析
// 2.将docker容器ip地址固定
func SetupCluster(conf *config.Config) {
	//fmt.Println(readClusterInfo())
	RedisCluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        readClusterInfo(),
		MaxRedirects: 8,
		PoolSize:     15,
		MinIdleConns: 10,
		Password:     conf.RedisConfig.Password,
		DialTimeout:  5 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  5 * time.Minute,
		OnConnect: func(conn *redis.Conn) error {
			fmt.Printf("conn=%v\n", conn)
			return nil
		},
	})
	fmt.Println(RedisCluster.Ping().Result())
	fmt.Println("connect to redis-cluster....")
}

func GetKey(key string) (string, error) {
	return RedisCluster.Get(key).Result()
}

// set expires为0时，表示永久性存储
func SetKey(key string, value interface{}, expires int) error {
	_, err := RedisCluster.Set(key, value, time.Duration(expires)*time.Second).Result()
	return err
}

func Keys(prefix string) ([]string, error) {
	return RedisCluster.Keys(prefix).Result()
}

// del
func DelKey(key string) error {
	// rds := RedisPool.Get()
	// defer rds.Close()
	// _, err := rds.Do("DEL", key)
	_, err := RedisCluster.Del(key).Result()
	return err
}
func DelKeys(keys ...string) error {
	_, err := RedisCluster.Do("DEL", keys).Result()
	return err
}
func RPop(key string, vlaue interface{}, expires int) error {
	return nil
}

// lrange
func LRange(key string, start, stop int64) ([]string, error) {
	//rds := RedisPool.Get()
	//defer rds.Close()
	return RedisCluster.LRange(key, start, stop).Result()
}

func LPush(key string, v ...interface{}) (int64, error) {
	return RedisCluster.LPush(key, v...).Result()
}

// lpop
func LPop(key string) (string, error) {
	// rds := RedisPool.Get()
	// defer rds.Close()
	return RedisCluster.LPop(key).Result()
	//return redis.String(rds.Do("LPOP", key))
}

// ExistsKey
func ExistsKey(key string) (bool, error) {
	i, err := RedisCluster.Exists(key).Result()
	return i == 0, err
}

// ttl 返回剩余时间
func TTLKey(key string) (int64, error) {
	ttl, err := RedisCluster.TTL(key).Result()
	return int64(ttl), err
}

// incr 自增
func Incr(key string) (int64, error) {
	return RedisCluster.Incr(key).Result()
}

// Decr 自减
func Decr(key string) (int64, error) {
	return RedisCluster.Decr(key).Result()
}

// mset 批量写入 rds.Do("MSET", "ket1", "value1", "key2","value2")
func MsetKey(key_value ...interface{}) error {
	// rds := RedisPool.Get()
	// defer rds.Close()
	// _, err := rds.Do("MSET", key_value...)
	_, err := RedisCluster.MSet(key_value...).Result()
	return err
}

func MsetKeyExpire(expire int, key_value ...interface{}) error {
	pi := RedisCluster.Pipeline()
	pi.MSet(key_value...)
	for i, v := range key_value {
		if i%2 == 0 {
			//keys = append(keys, fmt.Sprintf("%v", v))
			pi.Expire(fmt.Sprintln("%v", v), time.Second*time.Duration(expire))
		}
	}
	_, err := pi.Exec()
	return err
}

// mget  批量读取 mget key1, key2, 返回map结构
func MgetKey(keys ...string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	results, _ := RedisCluster.MGet(keys...).Result()

	for i := 0; i < len(keys); i++ {
		resultMap[keys[i]] = results[i]
	}
	return resultMap
}
