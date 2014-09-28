package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/alexcesaro/log/stdlog"
	"github.com/gorilla/mux"
)

var logger = stdlog.GetFromFlags()

var DBPREFIX = "perf"

var Storage = MongoHandler{}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Infof("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := os.Getenv("GOPATH") + "/src/github.com/pavel-paulau/perfkeeper/app"
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(app)))
	http.Handle("/static/", fileHandler)

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", ListDatabases).Methods("GET")
	r.HandleFunc("/{db}", ListSources).Methods("GET")
	r.HandleFunc("/{db}/{source}", ListMetrics).Methods("GET")
	r.HandleFunc("/{db}/{source}", AddSamples).Methods("POST")
	r.HandleFunc("/{db}/{source}/{metric}", GetRawValues).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/summary", GetSummary).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/linechart", GetLineChart).Methods("GET")
	http.Handle("/", r)

	fmt.Println("\n\t:-:-: perfkeeper :-:-:\t\t\tserving http://0.0.0.0:8080/\n")

	err := Storage.Init()
	if err != nil {
		os.Exit(1)
	}

	logger.Critical(http.ListenAndServe("0.0.0.0:8080", Log(http.DefaultServeMux)))
}
