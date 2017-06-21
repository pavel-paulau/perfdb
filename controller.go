package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Controller struct {
	storage *perfDB
}

func newController(storage *perfDB) *Controller {
	return &Controller{storage}
}

func newConn(rw http.ResponseWriter, r *http.Request) (*restHandler, error) {
	conn := &restHandler{rw, r}
	err := conn.open()
	return conn, err
}

func (c *Controller) listDatabases(rw http.ResponseWriter, r *http.Request) {
	conn, err := newConn(rw, r)
	if err != nil {
		logger.Critical(err)
		return
	}

	databases, err := c.storage.listDatabases()
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}

	conn.writeJSON(databases)
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

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}
	sources, err := c.storage.listSources(dbname)
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}
	conn.writeJSON(sources)
}

func (c *Controller) listMetrics(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	metrics, err := c.storage.listMetrics(dbname, source)
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}
	conn.writeJSON(metrics)
}

func (c *Controller) getRawValues(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	values, err := c.storage.getRawValues(dbname, source, metric)
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}
	conn.writeJSON(values)
}

func (c *Controller) getSummary(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	values, err := c.storage.getSummary(dbname, source, metric)
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}
	conn.writeJSON(values)
}

func (c *Controller) addSamples(rw http.ResponseWriter, r *http.Request) {
	var tsNano int64
	if timestamps, ok := r.URL.Query()["ts"]; ok {
		tsNano = parseTimestamp(timestamps[0])
	} else {
		tsNano = time.Now().UnixNano()
	}

	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]

	conn, err := newConn(rw, r)

	samples, err := conn.readJSON()
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 400)
		return
	}

	go func() {
		for metric, value := range samples.(map[string]interface{}) {
			sample := Sample{tsNano, value.(float64)}
			c.storage.addSample(dbname, source, metric, sample)
		}
	}()

	conn.writeJSON(map[string]string{"status": "ok"})
}

func (c *Controller) getHeatMap(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	hm, err := c.storage.getHeatMap(dbname, source, metric)
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}
	conn.writeJSON(hm)
}

func (c *Controller) getHistogram(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	values, err := c.storage.getHistogram(dbname, source, metric)
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 500)
		return
	}
	conn.writeJSON(values)
}

func (c *Controller) getHeatMapSVG(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
	source := vars["source"]
	metric := vars["metric"]

	var title string
	if titles, ok := r.URL.Query()["label"]; ok {
		title = titles[0]
	} else {
		title = metric
	}

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		return
	}

	hm, err := c.storage.getHeatMap(dbname, source, metric)
	if err != nil {
		logger.Critical(err)
		return
	}

	rw.Header().Set("Content-Type", "image/svg+xml")
	generateSVG(rw, hm, title)
}
