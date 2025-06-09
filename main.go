package main

import (
	"sync"
)

var logs sync.Map

func main() {
	go Index()
	go mitmproxy()
	scan()
}
