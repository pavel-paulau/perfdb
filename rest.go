package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/tebeka/nrsc"
	"github.com/gorilla/mux"
)

type httpResponse struct {
	Raw interface{}
}

func (r httpResponse) String() (s string) {
	b, err := json.Marshal(r.Raw)
	if err != nil {
		b, err = json.Marshal(map[string]string{
			"error": fmt.Sprint("", err),
		})
	}
	s = string(b)
	return
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", listDatabases).Methods("GET")
	r.HandleFunc("/{db}", listSources).Methods("GET")
	r.HandleFunc("/{db}/{source}", listMetrics).Methods("GET")
	r.HandleFunc("/{db}/{source}", addSamples).Methods("POST")
	r.HandleFunc("/{db}/{source}/{metric}", getRawValues).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/summary", getSummary).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/linechart", getLineChart).Methods("GET")

	return r
}

func internalError(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(rw, "Internal Server Error")
}

func validJSON(rw http.ResponseWriter, data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, httpResponse{data})
}

func validHTML(rw http.ResponseWriter, content string) {
	rw.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(rw, content)
}

func listDatabases(rw http.ResponseWriter, r *http.Request) {
	databases, err := storage.listDatabases()
	if err != nil {
		internalError(rw)
		return
	}
	validJSON(rw, databases)
}

func stringInSlice(a string, array []string) bool {
	for _, b := range array {
		if b == a {
			return true
		}
	}
	return false
}

func checkDbExists(rw http.ResponseWriter, dbname string) error {
	if allDbs, err := storage.listDatabases(); !stringInSlice(dbname, allDbs) || err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "Not existing snapshot")
		return errors.New("Not existing snapshot")
	}
	return nil
}

func listSources(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]

	if err := checkDbExists(rw, dbname); err != nil {
		return
	}
	sources, err := storage.listCollections(dbname)
	if err != nil {
		internalError(rw)
		return
	}
	validJSON(rw, sources)
}

func listMetrics(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]

	if err := checkDbExists(rw, dbname); err != nil {
		return
	}

	metrics, err := storage.listMetrics(dbname, source)
	if err != nil {
		internalError(rw)
		return
	}
	validJSON(rw, metrics)
}

func getRawValues(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if err := checkDbExists(rw, dbname); err != nil {
		return
	}

	values, err := storage.findValues(dbname, source, metric)
	if err != nil {
		internalError(rw)
		return
	}
	validJSON(rw, values)
}

func getSummary(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if err := checkDbExists(rw, dbname); err != nil {
		return
	}

	values, err := storage.aggregate(dbname, source, metric)
	if err != nil {
		internalError(rw)
		return
	}
	validJSON(rw, values)
}

func readHTML(path string) (string, error) {
	var html nrsc.Resource
	if html = nrsc.Get(path); html == nil {
		return "", errors.New("cannot read HTML")
	}
	var htmlReader io.Reader
	var err error
	if htmlReader, err = html.Open(); err != nil {
		return "", err
	}
	var content []byte
	if content, err = ioutil.ReadAll(htmlReader); err != nil {
		return "", err
	}
	return string(content), nil
}

func getLineChart(rw http.ResponseWriter, r *http.Request) {
	content, err := readHTML("linechart.html")
	if err != nil {
		internalError(rw)
		return
	}
	validHTML(rw, content)
}

func addSamples(rw http.ResponseWriter, r *http.Request) {
	var tsNano int64
	if timestamps, ok := r.URL.Query()["ts"]; ok {
		tsNano = parseTimestamp(timestamps[0])
	} else {
		tsNano = time.Now().UnixNano()
	}
	ts := strconv.FormatInt(tsNano, 10)

	vars := mux.Vars(r)
	dbname := vars["db"]
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
		go storage.insertSample(dbname, source, sample)
	}
}
