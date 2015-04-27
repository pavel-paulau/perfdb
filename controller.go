package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Controller struct {
	storage Storage
	rest restHanlder
}

func newController(storage Storage) *Controller {
	return &Controller{storage, restHanlder{}}
}

func (c *Controller) listDatabases(rw http.ResponseWriter, r *http.Request) {
	databases, err := c.storage.listDatabases()
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, databases)
}

func stringInSlice(a string, array []string) bool {
	for _, b := range array {
		if b == a {
			return true
		}
	}
	return false
}

func (c *Controller) checkDbExists(dbname string) error {
	if allDbs, err := c.storage.listDatabases(); !stringInSlice(dbname, allDbs) || err != nil {
		return errors.New("not found")
	}
	return nil
}

func (c *Controller) listSources(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]

	if err := c.checkDbExists(dbname); err != nil {
		c.rest.propagateError(rw, err, 404)
		return
	}
	sources, err := c.storage.listSources(dbname)
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, sources)
}

func (c *Controller) listMetrics(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]

	if err := c.checkDbExists(dbname); err != nil {
		c.rest.propagateError(rw, err, 404)
		return
	}

	metrics, err := c.storage.listMetrics(dbname, source)
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, metrics)
}

func (c *Controller) getRawValues(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if err := c.checkDbExists(dbname); err != nil {
		c.rest.propagateError(rw, err, 404)
		return
	}

	values, err := c.storage.getRawValues(dbname, source, metric)
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, values)
}

func (c *Controller) getSummary(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if err := c.checkDbExists(dbname); err != nil {
		c.rest.propagateError(rw, err, 404)
		return
	}

	values, err := c.storage.getSummary(dbname, source, metric)
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, values)
}

func (c *Controller) addSamples(rw http.ResponseWriter, r *http.Request) {
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
	decoder := json.NewDecoder(c.rest.readRequest(r))
	err := decoder.Decode(&samples)
	if err != nil {
		c.rest.propagateError(rw, err, 400)
		return
	}

	for metric, value := range samples {
		sample := Sample{ts, value.(float64)}
		go c.storage.addSample(dbname, source, metric, sample)
	}
}

func (c *Controller) getHeatMap(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if err := c.checkDbExists(dbname); err != nil {
		c.rest.propagateError(rw, err, 404)
		return
	}

	values, err := c.storage.getHeatMap(dbname, source, metric)
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, values)
}

func (c *Controller) getHistogram(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	if err := c.checkDbExists(dbname); err != nil {
		c.rest.propagateError(rw, err, 404)
		return
	}

	values, err := c.storage.getHistogram(dbname, source, metric)
	if err != nil {
		c.rest.propagateError(rw, err, 500)
		return
	}
	c.rest.validJSON(rw, values)
}
