package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
)

var (
	logger        *golog.Logger
	address, path *string
)

func requestLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Infof("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

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

	// RESTful API and HTML pages
	http.Handle("/", newRouter(controller))

	// Banner and launcher
	banner := fmt.Sprintf("\n\t:-:-: perfdb :-:-:\t\t\tserving http://%s/\n", *address)
	fmt.Println(banner)
	logger.Critical(http.ListenAndServe(*address, requestLog(http.DefaultServeMux)))
}
