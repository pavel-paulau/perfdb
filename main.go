package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/alexcesaro/log/stdlog"
)

var logger = stdlog.GetFromFlags()

var storage storageHandler

func requestLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Infof("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Storage initialization
	var err error
	if storage, err = newMongoHandler(); err != nil {
		os.Exit(1)
	}

	// Static assets
	app := os.Getenv("GOPATH") + "/src/github.com/pavel-paulau/perfkeeper/app"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(app))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(app)))

	// RESTful API
	http.Handle("/", newRouter())

	// Banner and launcher
	fmt.Println("\n\t:-:-: perfkeeper :-:-:\t\t\tserving http://0.0.0.0:8080/\n")
	logger.Critical(http.ListenAndServe("0.0.0.0:8080", requestLog(http.DefaultServeMux)))
}
