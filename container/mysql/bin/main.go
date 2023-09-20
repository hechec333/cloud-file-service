package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"

	"github.com/go-ini/ini"
)

var MntPath = "data"

type Compose struct {
	Version string `ini:"version"`
	Image   string `ini:"image"`
}
type MysqlMaster struct {
	Label    string
	HostName string `ini:"hostname"`
	Port     string `ini:"port"`
	Volume   string `ini:"volume"`
}
type MysqlSlave struct {
	Label    string
	HostName string `ini:"hostname"`
	Port     string `ini:"port"`
	Volume   string `ini:"volume"`
}
type Cu struct {
	UserName string `ini:"username"`
	PassWord string `ini:"password"`
}
type Mysql struct {
	SyncThreads int    `ini:"syncthreads"`
	HalfSync    string `ini:"halfsync"`
}
type ComposeConfig struct {
	Compose
	MysqlMaster
	Cu
	Mysql
	Slaves []MysqlSlave
}

func LoadConfig() (*ComposeConfig, error) {
	Cfg, _ := ini.Load("service.ini")
	config := ComposeConfig{}
	secs := Cfg.Sections()
	for _, v := range secs {
		var err error
		if v.Name() == "compose" {
			err = v.MapTo(&config.Compose)

		} else if v.Name() == "master" {
			err = v.MapTo(&config.MysqlMaster)
			config.MysqlMaster.Label = "master"
		} else if v.Name() == "DEFAULT" {
			continue
		} else if v.Name() == "cu" {
			err = v.MapTo(&config.Cu)
		} else if v.Name() == "mysql" {
			err = v.MapTo(&config.Mysql)
		} else {
			slave := MysqlSlave{}
			slave.Label = v.Name()
			err = v.MapTo(&slave)
			if err != nil {
				goto e
			}
			config.Slaves = append(config.Slaves, slave)
		}
	e:
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return &config, nil
}
func isFileExist(pt string) bool {
	if _, err := os.Stat(pt); err != nil {
		return os.IsNotExist(err)
	}
	return true
}
func MakeMysqlContainer() error {

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	err = ComposeFile("../", cfg)
	if err != nil {
		return err
	}

	return ExecGenScript(cfg)
}
func ExecGenScript(cfg *ComposeConfig) error {
	cmd := "./gen.sh"
	args := []string{}
	for _, v := range cfg.Slaves {
		args = append(args, v.Label)
	}
	rcmd := exec.Command(cmd, args...)
	bys, err := rcmd.Output()
	if err != nil {
		log.Println(err, "what:", bys)
		return err
	}
	return nil
}
func ComposeFile(outpath string, cfg *ComposeConfig) error {
	var path = outpath + "docker-compose.yml"
	tmpl, err := template.New("docker-compose.tmpl").ParseFiles("./docker-compose.tmpl")
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

	if isFileExist(path) {
		os.Remove(path)
	}

	file, _ := os.Create(path)

	file.WriteString(buf.String())
	return nil
}
func exec_cu() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	args := []string{}
	for _, v := range cfg.Slaves {
		args = append(args, v.Label)
	}
	fmt.Println("executing exec_master.sh")
	cmd := "./exec_master.sh"
	rcmd1 := exec.Command(cmd, cfg.Cu.UserName, cfg.Cu.PassWord)
	bys, err := rcmd1.Output()
	fmt.Println(string(bys))
	if err != nil {
		fmt.Println(err)
		return errors.New("fail to exectute 'exec_master.sh' " + err.Error())
	}
	fmt.Println("executing query_slave.sh")
	cmd = "./query_slave.sh"
	rcmd2 := exec.Command(cmd, args...)
	bys, err = rcmd2.Output()
	fmt.Println(string(bys))
	if err != nil {
		fmt.Println(rcmd2.StderrPipe())
		return errors.New("fail to execute 'query_slave.sh' " + err.Error())
	}

	return DoCleanTmpFile()
}

func MustLoadMysqlOption() {

}

func DoCleanTmpFile() error {
	cmd := "rm"
	c := exec.Command(cmd, "rf", "../tmp/*")
	c.Run()
	return nil
}
func main() {
	fmt.Println(os.Getwd())
	if len(os.Args) < 1 {
		fmt.Println("enter at lease one arg [compose,exec-cu]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "compose":
		err := MakeMysqlContainer()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	case "exec-cu":
		err := exec_cu()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("command: ", os.Args[1], " not supported")
		os.Exit(1)
	}
	//MakeMysqlContainer()
	//exec_cu()
}
