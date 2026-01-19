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
		if len(args) != 2 {
			println("Wrong number of arguments. Be sure to run\n\n\tgitport start <port>")
			return
		}
		server.Start(args[2])
	case "init":
		if len(args) != 1 {
			println("gitport init doesn't take arguments")
			return
		}
		server.Init()
	case "help":
		println("Help coming soon")
	}

}
