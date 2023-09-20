package main

import (
	"fmt"
	"orm/config"
	"orm/dao/cache"
	"orm/dao/db"
	"orm/router"
)

func main() {

	config := config.GetConfig()

	cache.SetupCluster(config)
	db.Setup(config)

	address := ":8080"

	engine := router.ImportRoutes()

	if err := engine.Run(address); err != nil {
		panic(err)
	}

	fmt.Println("server running at ", address)
}
