package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"bitbucket.org/tebeka/nrsc"
	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
	"github.com/davecheney/profile"
)

var (
	logger        *golog.Logger
	cpu           *bool
	address, path *string
)

func requestLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Infof("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

func init() {
	runtime.GOMAXPROCS(2)

	address = flag.String("address", "127.0.0.1:8080", "serve requests to this host[:port]")
	path = flag.String("path", "/tmp/perfdb", "PerfDB data directory")
	cpu = flag.Bool("cpu", false, "Enable CPU profiling")
	flag.Parse()

	logger = golog.New(os.Stdout, log.Info)
}

func main() {
	// Optionally enable CPU profiling
	if *cpu {
		cfg := profile.Config{
			ProfilePath: ".",
			CPUProfile:  true,
		}
		defer profile.Start(&cfg).Stop()
	}

	// Database handler
	var err error
	var storage Storage
	if storage, err = newPerfDB(*path); err != nil {
		os.Exit(1)
	}

	// Controller
	controller := newController(storage)

	// RESTful API and HTML pages
	http.Handle("/", newRouter(controller))

	// Static assets
	nrsc.Handle("/static/")

	// Banner and launcher
	banner := fmt.Sprintf("\n\t:-:-: perfkeeper :-:-:\t\t\tserving http://%s/\n", *address)
	fmt.Println(banner)
	logger.Critical(http.ListenAndServe(*address, requestLog(http.DefaultServeMux)))
}
