package main

import (
	"flag"
	"fmt"
	"log"

	"./broker"
	"./client"
	"./config"
	"./server"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var configFile string
	flag.StringVar(&configFile, "config", "client.json", "The configuration JSON file path")
	flag.Parse()

	if len(configFile) == 0 {
		panic("`config` parameter is missing")
	}

	err := config.Load(configFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config file %v : %v", configFile, err.Error()))
	}

	role := config.GetRole()
	log.Println("Starting detour proxy instance as", role, "role...")

	if role == config.RoleClient {
		err := client.Run(config.GetHttpPort(), config.GetSocksPort(), config.GetUrl())
		if err != nil {
			panic(err)
		}
	} else if role == config.RoleBroker {
		err := broker.Run(config.GetHttpPort())
		if err != nil {
			panic(err)
		}
	} else if role == config.RoleServer {
		err := server.Run(config.GetUrl())
		if err != nil {
			panic(err)
		}
	}

}
