package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		storage.InsertSample(DBPREFIX+db, source, sample)
	}
}

func main() {
	storage.Init()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", ListDatabases).Methods("GET")
	r.HandleFunc("/{db}", ListSources).Methods("GET")
	r.HandleFunc("/{db}/{source}", ListMetrics).Methods("GET")
	r.HandleFunc("/{db}/{source}", AddSamples).Methods("POST")
	r.HandleFunc("/{db}/{source}/{metric}", GetRawValues).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/summary", GetSummary).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
