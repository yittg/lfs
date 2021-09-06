package main

import (
	"github.com/yittg/lfs"
	"github.com/yittg/lfs/cmd"
)

func main() {
	cmd.SetLogger()

	server := lfs.NewServer(&lfs.Configuration{})
	server.Run()
}
