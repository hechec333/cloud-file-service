package config

import "github.com/go-ini/ini"

type Secret struct {
	AccessSecret string `ini:"AccessSecret"`
	AccessExpire int64  `ini:"AccessExpire"`
}

type RedisConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	Password string `ini:"password"`
}
type MysqlConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	Password string `ini:"password"`
	User     string `ini:"user"`
	Db       string `ini:"db"`
}

type CosConfig struct {
	SecretId     string `ini:"COS_SECRET_ID"`
	SecretKey    string `ini:"COS_SECRET_KEY"`
	BucketName   string `ini:"COS_BUCKET_NAME"`
	BucketRegiom string `ini:"COS_BUCKET_REGION"`
}

type OssConfig struct {
	AccessId     string `ini:"ACCESS_KEY_ID"`
	AccessSecret string `ini:"ACCESS_KEY_SECRET"`
	EndPoint     string `ini:"END_POINT"`
	BucketName   string `ini:"BUCKET_NAME"`
}

type HdfsConfig struct {
	HdfsMaster string `ini:"HDFS_Master"`
}

type Config struct {
	Secret
	RedisConfig
	MysqlConfig
	CosConfig
	OssConfig
	HdfsConfig
}

var conf *Config = nil

func GetConfig() *Config {
	if conf == nil {
		conf = mustloadConfig()
	}
	return conf
}

func mustloadConfig() *Config {
	var err error
	var cfg Config
	var Cfg *ini.File
	Cfg, err = ini.Load("config/service.ini")
	for _, v := range Cfg.Sections() {
		switch v.Name() {
		case "secret":
			err = v.MapTo(&cfg.Secret)
			if err != nil {
				panic(err)
			}
		case "redis":
			err = v.MapTo(&cfg.RedisConfig)
			if err != nil {
				panic(err)
			}
		case "mysql":
			err = v.MapTo(&cfg.MysqlConfig)
			if err != nil {
				panic(err)
			}
		case "oss":
			err = v.MapTo(&cfg.OssConfig)
			if err != nil {
				panic(err)
			}
		case "cos":
			err = v.MapTo(&cfg.CosConfig)
			if err != nil {
				panic(err)
			}
		case "hdfs":
			err = v.MapTo(&cfg.HdfsConfig)
			if err != nil {
				panic(err)
			}
		default:
			continue
		}
	}
	return &cfg
}
