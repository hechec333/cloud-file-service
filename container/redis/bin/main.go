package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-ini/ini"
)

type RedisCluster struct {
	Password string `ini:"password"`
	PortSeed string `ini:"portSeed"`
}
type ComposeConfig struct {
	Version string `ini:"version"`
	Image   string `ini:"image"`
	Gateway string `ini:"gateway"`
}
type MasterConfig struct {
	HostName string `ini:"hostname"`
	PassWord string `ini:"password"`
	Volume   string `ini:"volume"`
	Port     string `ini:"port"`
	Label    string
}

type SlaveConfig struct {
	HostName string `ini:"hostname"`
	PassWord string `ini:"password"`
	Volume   string `ini:"volume"`
	Port     string `ini:"port"`
	Label    string
}

type RedisNodeConfig struct {
	VolumeDir string `ini:"mountDir"`
	Label     string
}
type RedisConfig struct {
	RedisCluster
	ComposeConfig
	NodeConfig []RedisNodeConfig
}

func MustLoadConfig() *RedisConfig {
	Cfg, _ := ini.Load("service.ini")
	config := RedisConfig{}
	desc := Cfg.Sections()
	for _, v := range desc {
		var err error

		switch v.Name() {
		case "DEFAULT":
			continue
		case "compose":
			err = v.MapTo(&config.ComposeConfig)
			if err != nil {
				panic(err)
			}
		case "redis-cluster":
			err = v.MapTo(&config.RedisCluster)
			if err != nil {
				panic(err)
			}
		default:
			if strings.Contains(v.Name(), "redis_node") {
				m := RedisNodeConfig{}
				m.Label = v.Name()[strings.Index(v.Name(), "redis_"):]
				err = v.MapTo(&m)
				if err != nil {
					panic(err)
				}
				config.NodeConfig = append(config.NodeConfig, m)
			}

		}
	}
	return &config
}
func isFileOrDicExist(pt string) bool {
	if _, err := os.Stat(pt); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
func CreateVolumeDir(outpath string, cfg *RedisConfig) error {
	path := outpath + "/data"
	if !isFileOrDicExist(path) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			return err
		}
	}
	for _, v := range cfg.NodeConfig {
		if isFileOrDicExist(path + "/" + v.VolumeDir) {
			continue
		}
		err := os.Mkdir(path+"/"+v.VolumeDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateEtcDir(outpath string, cfg *RedisConfig) error {
	path := outpath + "/etc"
	if !isFileOrDicExist(path) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			return err
		}
	}
	for _, v := range cfg.NodeConfig {
		if isFileOrDicExist(path + "/" + v.Label) {
			continue
		}
		err := os.Mkdir(path+"/"+v.Label, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func ComposeFile(outpath string, cfg *RedisConfig) error {
	var path = outpath + "docker-compose.yml"
	tmpl, err := template.New("docker-compose.tmpl").Funcs(template.FuncMap{
		"getPort": func(arg int) string {
			number, _ := strconv.Atoi(cfg.RedisCluster.PortSeed)
			return strconv.Itoa(number + arg)
		},
		"getIpAddr": func(arg int) string {
			tokens := strings.Split(cfg.ComposeConfig.Gateway, ".")
			seed, _ := strconv.Atoi(tokens[3])
			tokens[3] = strconv.Itoa(seed + 2 + arg)
			return strings.Join(tokens, ".")
		},
	}).ParseFiles("./docker-compose.tmpl")

	if err != nil {
		log.Println(err)
		return err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, *cfg)

	if err != nil {
		log.Println(err)
		return err
	}

	if isFileOrDicExist(path) {
		os.Remove(path)
	}

	file, _ := os.Create(path)

	file.WriteString(buf.String())
	return nil

}

func MustGenRedisMountDir(cfg *RedisConfig) {

	var err error
	err = CreateVolumeDir("..", cfg)
	if err != nil {
		panic(err)
	}

	err = CreateEtcDir("..", cfg)
	if err != nil {
		panic(err)
	}
}

func getRedisIpAddr(index int, cfg *RedisConfig) string {
	tokens := strings.Split(cfg.ComposeConfig.Gateway, ".")
	seed, _ := strconv.Atoi(tokens[3])
	tokens[3] = strconv.Itoa(seed + 2 + index)
	return strings.Join(tokens, ".")
}

func MustGenRedisConfigFile(cfg *RedisConfig) {

	for i, v := range cfg.NodeConfig {
		cmd := exec.Command("./envsub.sh")
		cmd.Env = append(cmd.Env, "PortSeed="+cfg.RedisCluster.PortSeed, "RedisPassword="+cfg.RedisCluster.Password, "HostName="+v.Label, "IpAddr="+getRedisIpAddr(i, cfg))
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}

func DockerHost2IpAddress(cfg *RedisConfig) []string {
	var endpoints []string
	for _, v := range cfg.NodeConfig {
		cmd := exec.Command("docker", "inspect", "--format", "'{{ .NetworkSettings.Networks.redis_redis_network.IPAddress }}'", v.Label)
		bys, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		endpoints = append(endpoints, string(bys[1:len(bys)-2])+":"+cfg.PortSeed) //trim '\n'
	}
	return endpoints
}

func MakeRedisCluster() error {
	cfg := MustLoadConfig()
	if cfg == nil {
		return errors.New("load service.ini file error")
	}
	MustGenRedisMountDir(cfg)
	MustGenRedisConfigFile(cfg)
	return ComposeFile("../", cfg)
}
func MakeRedisClusterLink() error {
	cfg := MustLoadConfig()
	cmd := exec.Command("./envtmpl.sh")
	// for i := 0; i < len(cfg.NodeConfig); i++ {
	// 	if i != len(cfg.NodeConfig)-1 {
	// 		clusters += cfg.NodeConfig[i].Label + ":6379 "
	// 		continue
	// 	}
	// 	clusters += cfg.NodeConfig[i].Label + ":6379)"
	// }
	clusters := DockerHost2IpAddress(cfg)
	cmd.Env = append(cmd.Env, "RedisPassword="+cfg.RedisCluster.Password, "clusters="+strings.Join(clusters, " "))
	fmt.Println(cmd.Env)
	bys, err := cmd.Output()
	fmt.Println(string(bys))
	return err
}

func MakeRedisInfoOutput() error {
	args := []string{}
	conf := MustLoadConfig()
	seed, _ := strconv.Atoi(conf.RedisCluster.PortSeed)
	for i, _ := range conf.NodeConfig {
		args = append(args, strconv.Itoa(seed+i))
	}
	cmd := exec.Command("./redis-var.sh", args...)
	cmd.Env = append(cmd.Env, "host=127.0.0.1", "password="+conf.RedisCluster.Password)
	err := cmd.Run()
	return err
}

func main() {

	if len(os.Args) < 2 {
		panic("unsuppoted args len")
	}

	switch os.Args[1] {
	case "gen":
		err := MakeRedisCluster()
		if err != nil {
			panic(err)
		}
	case "link":
		err := MakeRedisClusterLink()
		if err != nil {
			panic(err)
		}
	case "var":
		err := MakeRedisInfoOutput()
		if err != nil {
			panic(err)
		}
	default:
		panic("unknown argv:" + os.Args[1])
	}
	//MakeRedisCluster()
	// err := MakeRedisClusterLink()
	// if err != nil {
	// 	panic(err)
	// }
}
