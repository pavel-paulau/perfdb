package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/alexcesaro/log/stdlog"
)

var Logger = stdlog.GetFromFlags()

var Storage StorageHandler

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		Logger.Infof("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Storage initialization
	var err error
	if Storage, err = NewMongoHandler(); err != nil {
		os.Exit(1)
	}

	// Static assets
	app := os.Getenv("GOPATH") + "/src/github.com/pavel-paulau/perfkeeper/app"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(app))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(app)))

	// RESTful API
	http.Handle("/", NewRouter())

	// Banner and launcher
	fmt.Println("\n\t:-:-: perfkeeper :-:-:\t\t\tserving http://0.0.0.0:8080/\n")
	Logger.Critical(http.ListenAndServe("0.0.0.0:8080", Log(http.DefaultServeMux)))
}
