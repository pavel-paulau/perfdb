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

func (c *Controller) checkDbExists(dbname string) error {
	status, err := c.storage.isExist(dbname)
	if err != nil {
		return err
	}
	if !status {
		return errors.New("not found")
	}
	return nil
}

func (c *Controller) listMetrics(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	metrics, err := c.storage.listMetrics(dbname)
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
	metric := vars["metric"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	values, err := c.storage.getRawValues(dbname, metric)
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
	metric := vars["metric"]

	conn, err := newConn(rw, r)

	if err := c.checkDbExists(dbname); err != nil {
		logger.Critical(err)
		conn.writeError(err, 404)
		return
	}

	values, err := c.storage.getSummary(dbname, metric)
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

	conn, err := newConn(rw, r)

	samples, err := conn.readJSON()
	if err != nil {
		logger.Critical(err)
		conn.writeError(err, 400)
		return
	}

	for metric, value := range samples.(map[string]interface{}) {
		sample := Sample{tsNano, value.(float64)}
		if err := c.storage.addSample(dbname, metric, sample); err != nil {
			conn.writeError(err, 500)
			return
		}
	}

	conn.writeJSON(map[string]string{"status": "ok"})
}

func (c *Controller) getHeatMapSVG(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["db"]
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

	hm, err := c.storage.getHeatMap(dbname, metric)
	if err != nil {
		logger.Critical(err)
		return
	}

	rw.Header().Set("Content-Type", "image/svg+xml")
	generateSVG(rw, hm, title)
}
