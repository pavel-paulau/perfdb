package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Response struct {
	Raw interface{}
}

func (r Response) String() (s string) {
	b, err := json.Marshal(r.Raw)
	if err != nil {
		b, err = json.Marshal(map[string]string{
			"error": fmt.Sprint("", err),
		})
	}
	s = string(b)
	return
}

var DBPREFIX = "perf"

var storage = MongoHandler{}

func ListDatabases(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, Response{storage.ListDatabases()})
}

func ListSources(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]

	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, Response{storage.ListCollections(DBPREFIX + db)})
}

func ListMetrics(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]

	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, Response{storage.ListMetrics(DBPREFIX+db, source)})
}

func GetRawValues(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	rw.Header().Set("Content-Type", "application/json")
	values := storage.FindValues("perf"+db, source, metric)
	fmt.Fprint(rw, Response{values})
}

func GetSummary(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	rw.Header().Set("Content-Type", "application/json")
	values := storage.Aggregate("perf"+db, source, metric)
	fmt.Fprint(rw, Response{values})
}

func GetLineChart(rw http.ResponseWriter, r *http.Request) {
	app := os.Getenv("GOPATH") + "/src/github.com/pavel-paulau/perfkeeper/"

	content, _ := ioutil.ReadFile(app + "app/linechart.html")
	rw.Header().Set("Content-Type", "text/html")
	fmt.Fprint(rw, string(content))
}

func AddSamples(rw http.ResponseWriter, r *http.Request) {
	tsInt := time.Now().UnixNano()
	ts := strconv.FormatInt(tsInt, 10)

	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]

	var samples map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&samples)
	if err != nil {
		log.Fatalln(err)
	}

	for m, v := range samples {
		sample := map[string]interface{}{
			"ts": ts,
			"m":  m,
			"v":  v,
		}
		go storage.InsertSample(DBPREFIX+db, source, sample)
	}
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	storage.Init()

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

	fmt.Println("\n\t:-:-: perfkeeper :-:-:\t\t\tserving 0.0.0.0:8080\n")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", Log(http.DefaultServeMux)))
}
