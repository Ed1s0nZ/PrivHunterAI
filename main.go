package main

import (
	"sync"
)

var logs sync.Map

func main() {
	go mitmproxy()
	scan()
}
