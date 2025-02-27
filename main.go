package main

import (
	"sync"
	"yuequanScan/config"
)

var logs sync.Map

func main() {

	config.GetConfig()

	go Index()
	go mitmproxy()
	scan()
}
