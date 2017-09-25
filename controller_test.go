package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func removeStorage(pdb *perfDB) {
	os.RemoveAll(pdb.baseDir)
}

func newTmpStorage() (*perfDB, error) {
	var tmpDir string
	var err error

	if tmpDir, err = ioutil.TempDir("", ""); err != nil {
		return nil, err
	}

	var storage *perfDB
	if storage, err = newPerfDB(tmpDir); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(storage, removeStorage)
	return storage, nil
}

func TestListDatabases(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "[]", rw.Body.String())
}

func TestAddSample(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "{\"status\":\"ok\"}", rw.Body.String())
}

func TestAddBadJSON(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0,}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestListMetrics(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"latency_set\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "[\"latency_set\"]", rw.Body.String())
}

func TestListMetricsMissingDB(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/missing", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestGetRawValues(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515",
		bytes.NewBufferString("{\"cpu\":100510051005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411940890615",
		bytes.NewBufferString("{\"cpu\":0}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411940891708",
		bytes.NewBufferString("{\"cpu\":575.11}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t,
		"[[1411940889515,100510051005],[1411940890615,0],[1411940891708,575.11]]",
		rw.Body.String())
}

func TestGetRawValuesMissingDatabase(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("GET", "/database/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestGetRawValuesMissingMetric(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"cpu\":80}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/mem_used", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestGetSummary(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"cpu\":1005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"cpu\":75.11}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu/summary", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t,
		"{\"avg\":540.055,\"count\":2,\"max\":1005,\"min\":75.11,\"p50\":75.11,\"p80\":75.11,\"p90\":75.11,\"p95\":75.11,\"p99\":75.11,\"p99.9\":75.11}",
		rw.Body.String())
}

func TestGetSummaryMissingDatabase(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("GET", "/database/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestGetSummaryMissingMetric(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"cpu\":1005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/mem_used/summary", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestGetHeatmap(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515",
		bytes.NewBufferString("{\"cpu\":80}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411940890515",
		bytes.NewBufferString("{\"cpu\":75.11}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu/heatmap", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	contentType := rw.Header().Get("Content-Type")

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, contentType, "image/svg+xml")
	assert.Contains(t, rw.Body.String(), "<!-- Generated by SVGo -->")
	assert.Contains(t, rw.Body.String(), "Time elapsed, s")
}

func TestGetHeatmapLargeValues(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515",
		bytes.NewBufferString("{\"cpu\":10051005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411940989515",
		bytes.NewBufferString("{\"cpu\":10051005}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu/heatmap", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Contains(t, rw.Body.String(), "<!-- Generated by SVGo -->")
	assert.Contains(t, rw.Body.String(), "Time elapsed, m")
}

func TestGetHeatmapMinutes(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515",
		bytes.NewBufferString("{\"cpu\":1005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411940989515",
		bytes.NewBufferString("{\"cpu\":2005}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu/heatmap", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Contains(t, rw.Body.String(), "<!-- Generated by SVGo -->")
	assert.Contains(t, rw.Body.String(), "Time elapsed, m")
}

func TestGetHeatmapHours(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515",
		bytes.NewBufferString("{\"cpu\":0.000001}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411945889515",
		bytes.NewBufferString("{\"cpu\":0.000002}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu/heatmap", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Contains(t, rw.Body.String(), "<!-- Generated by SVGo -->")
	assert.Contains(t, rw.Body.String(), "Time elapsed, h")
}

func TestGetHeatmapWithLabel(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515",
		bytes.NewBufferString("{\"cpu\":0.001}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411941889515",
		bytes.NewBufferString("{\"cpu\":0.002}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu/heatmap?label=CPU", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Contains(t, rw.Body.String(), "<!-- Generated by SVGo -->")
}

func TestGetHeatmapMissingDatabase(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("GET", "/database/cpu/heatmap", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestGetHeatmapMissingMetric(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database",
		bytes.NewBufferString("{\"cpu\":1005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/mem_free/heatmap", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}
