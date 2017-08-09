package main

import (
	"flag"
	"os"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
)

var (
	logger        *golog.Logger
	address, path *string
)

func init() {
	address = flag.String("address", "127.0.0.1:8080", "serve requests to this host[:port]")
	path = flag.String("path", "data", "PerfDB data directory")
	flag.Parse()

	logger = golog.New(os.Stdout, log.Info)
}

func main() {
	// Database handler
	var err error
	var storage *perfDB
	if storage, err = newPerfDB(*path); err != nil {
		os.Exit(1)
	}

	// Controller
	controller := newController(storage)
	newRouter(controller).Run(*address)
}
