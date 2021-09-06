package main

import (
	"fmt"
	"os"
)

func main() {
	if err := NewLfsCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Command execute err, %v", err)
		os.Exit(1)
	}
}
