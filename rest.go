package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", ListDatabases).Methods("GET")
	r.HandleFunc("/{db}", ListSources).Methods("GET")
	r.HandleFunc("/{db}/{source}", ListMetrics).Methods("GET")
	r.HandleFunc("/{db}/{source}", AddSamples).Methods("POST")
	r.HandleFunc("/{db}/{source}/{metric}", GetRawValues).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/summary", GetSummary).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/linechart", GetLineChart).Methods("GET")

	return r
}

func ListDatabases(rw http.ResponseWriter, r *http.Request) {
	databases, err := Storage.ListDatabases()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, Response{databases})
	}
}

func ListSources(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]

	if sources, err := Storage.ListCollections(db); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, Response{sources})
	}
}

func ListMetrics(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]

	if metrics, err := Storage.ListMetrics(db, source); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, Response{metrics})
	}
}

func GetRawValues(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if values, err := Storage.FindValues(db, source, metric); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, Response{values})
	}
}

func GetSummary(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if values, err := Storage.Aggregate(db, source, metric); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, Response{values})
	}
}

func GetLineChart(rw http.ResponseWriter, r *http.Request) {
	app := os.Getenv("GOPATH") + "/src/github.com/pavel-paulau/perfkeeper/"

	if content, err := ioutil.ReadFile(app + "app/linechart.html"); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "text/html")
		fmt.Fprint(rw, string(content))
	}
}

func AddSamples(rw http.ResponseWriter, r *http.Request) {
	var tsNano int64
	if timestamps, ok := r.URL.Query()["ts"]; ok {
		tsNano = ParseTimestamp(timestamps[0])
	} else {
		tsNano = time.Now().UnixNano()
	}
	ts := strconv.FormatInt(tsNano, 10)

	vars := mux.Vars(r)
	db := vars["db"]
	source := vars["source"]

	var samples map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&samples)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "Cannot decode sample: %s\n", err)
		return
	}

	for m, v := range samples {
		sample := map[string]interface{}{
			"ts": ts,
			"m":  m,
			"v":  v,
		}
		go Storage.InsertSample(db, source, sample)
	}
}
