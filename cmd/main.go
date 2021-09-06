package main

import (
	"github.com/yittg/lfs"
)

func main() {
	setLogger()

	server := lfs.NewServer(&lfs.Configuration{})
	server.Run()
}
