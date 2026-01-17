package main

import (
	"os"

	"github.com/nim-sam/gitport/pkg/server"
)

/**
 * Entry point of the program
 */
func main() {
	args := os.Args

	switch args[1] {
	case "start":
		println("Starting server...")
		server.Start(args[2])
	}

}
