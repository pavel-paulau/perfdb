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

func TestListDatabasesPerfDb(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[]", rw.Body.String())
}

func TestAddSamplePerfDb(t *testing.T) {
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

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"status\":\"ok\"}", rw.Body.String())
}

func TestListMetricsPerfDb(t *testing.T) {
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

	req, _ = http.NewRequest("GET", "/database", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"cpu\"]", rw.Body.String())
}

func TestGetRawValuesPerfDb(t *testing.T) {
	var err error
	var storage *perfDB
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/database?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":1005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("POST", "/database?ts=1411940889515410775",
		bytes.NewBufferString("{\"cpu\":75.11}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	req, _ = http.NewRequest("GET", "/database/cpu", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t,
		"[[1411940889515410774,1005],[1411940889515410775,75.11]]",
		rw.Body.String())
}
