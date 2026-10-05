package main

import (
	"flag"
	"os"

	"github.com/charmbracelet/log"

	"github.com/nim-sam/gitport/internal/server"
)

func main() {

	host := flag.String("host", "0.0.0.0", "address to bind the ssh server to")
	port := flag.String("port", "23234", "port to bind the ssh server to")
	flag.Parse()

	server.GitportServer.Host = *host
	server.GitportServer.Port = *port

	// Start blocks until the listener exits or the process is interrupted
	if err := server.GitportServer.Start(); err != nil {
		log.Error("Could not run GitPort server", "error", err)
		os.Exit(1)
	}
}
