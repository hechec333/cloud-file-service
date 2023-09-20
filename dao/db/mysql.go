package db

import (
	"fmt"
	"orm/config"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Setup(conf *config.Config) {

	params := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
		conf.MysqlConfig.User,
		conf.MysqlConfig.Password,
		conf.MysqlConfig.Host+":"+conf.MysqlConfig.Port,
		conf.MysqlConfig.Db,
	)
	var err error
	DB, err = gorm.Open(mysql.Open(params), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NoLowerCase: true,
		},
	})
	if err != nil {
		panic(err)
	}
	db, _ := DB.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(30)
	fmt.Println("database init on port ", conf.MysqlConfig.Host+":"+conf.MysqlConfig.Port)
}
