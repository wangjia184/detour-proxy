package config

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

const RoleClient string = "client"
const RoleBroker string = "broker"
const RoleServer string = "server"

type Configuration struct {
	Role                string `json:"role"`
	HttpPort            int    `json:"httpPort"`
	SocksPort           int    `json:"socksPort"`
	Url                 string `json:"url"`
	GfwListUrl          string `json:"gfwListUrl"`
	SmartConnectTimeout int    `json:"smartConnectTimeout"`
}

var config Configuration

func Load(file string) error {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer, &config)
	if err != nil {
		return err
	}
	return nil
}

func GetRole() string {
	if config.Role != RoleClient &&
		config.Role != RoleBroker &&
		config.Role != RoleServer {
		panic("`role` must be 'client' / 'broker' / 'server'")
	}
	return config.Role
}

func GetHttpPort() uint16 {
	if config.HttpPort <= 0 ||
		config.HttpPort >= 65535 {
		panic("`httpPort` is invalid, please check your configuration file")
	}
	return uint16(config.HttpPort)
}

func GetSocksPort() uint16 {
	if config.SocksPort <= 0 ||
		config.SocksPort >= 65535 {
		panic("`socksPort` is invalid, please check your configuration file")
	}
	return uint16(config.SocksPort)
}

func GetUrl() string {
	if len(config.Url) == 0 {
		panic("`url` is missing, please check your configuration file")
	}
	if !strings.HasPrefix(config.Url, "ws://") &&
		!strings.HasPrefix(config.Url, "wss://") {
		panic("`url` must start with 'ws://' or 'wss://', please check your configuration file")
	}
	return config.Url
}

func GetGfwListUrl() string {
	if len(config.GfwListUrl) > 0 {
		if !strings.HasPrefix(config.GfwListUrl, "http://") &&
			!strings.HasPrefix(config.GfwListUrl, "https://") {
			panic("`gfwListUrl` must start with 'http://' or 'https://', please check your configuration file")
		}
	}

	return config.GfwListUrl
}

func GetSmartConnectTimeout() int64 {
	return int64(config.SmartConnectTimeout)
}
