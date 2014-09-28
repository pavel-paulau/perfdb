package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/alexcesaro/log/stdlog"
	"github.com/gorilla/mux"
)

var Logger = stdlog.GetFromFlags()

var Storage *MongoHandler

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
	if Storage, err = NewStorage(); err != nil {
		os.Exit(1)
	}

	// Static assets
	app := os.Getenv("GOPATH") + "/src/github.com/pavel-paulau/perfkeeper/app"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(app))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(app)))

	// RESTful API
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

	// Banner and launcher
	fmt.Println("\n\t:-:-: perfkeeper :-:-:\t\t\tserving http://0.0.0.0:8080/\n")
	Logger.Critical(http.ListenAndServe("0.0.0.0:8080", Log(http.DefaultServeMux)))
}
