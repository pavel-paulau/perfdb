package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MongoDb/TokuMX Unit Tests

func TestListDatabasesMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"snapshot\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListSourcesMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listSources", "snapshot").Return([]string{"source"}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"source\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

var ErrTest = errors.New("fake test error")

func TestListSourcesErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listSources", "snapshot").Return([]string{}, ErrTest)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListSourcesWrongSnapshotMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshotx", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 404, rw.Code)
	assert.Equal(t, "{\"error\":\"not found\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listMetrics", "snapshot", "source").Return([]string{"cpu_user", "cpu_sys"}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"cpu_user\",\"cpu_sys\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listMetrics", "snapshot", "source").Return([]string{}, ErrTest)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsWrongSnapshotMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshotx/source", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 404, rw.Code)
	assert.Equal(t, "{\"error\":\"not found\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestGetRawValuesMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getRawValues",
		"snapshot", "source", "cpu").Return(map[string]float64{"1411534805453497432": 100}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"1411534805453497432\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestGetRawValuesErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getRawValues",
		"snapshot", "source", "cpu").Return(map[string]float64{}, ErrTest)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestAddSampleMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("addSample", "snapshot", "source", "cpu",
		Sample{"1411940889515410774", 99.0}).Return(nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestAddSampleErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("addSample", "snapshot", "source", "cpu",
		Sample{"1411940889515410774", 99.0}).Return(ErrTest)

	controller := newController(storageMock)

	req, err := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	if err != nil {
		log.Fatal(err)
	}
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestAddBadSampleMongo(t *testing.T) {
	storageMock := new(mongoMock)
	controller := newController(storageMock)

	req, _ := http.NewRequest("POST", "/snapshot/source",
		bytes.NewBufferString(""))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 400, rw.Code)
	assert.Equal(t, "{\"error\":\"EOF\"}", rw.Body.String())
}

func TestSummaryMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getSummary",
		"snapshot", "source", "cpu").Return(map[string]interface{}{"mean": 100}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"mean\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestSummaryErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getSummary", "snapshot", "source", "cpu").Return(
		map[string]interface{}{}, ErrTest)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestHeatMapMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getHeatMap", "snapshot", "source", "cpu").Return(
		newHeatMap(), nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/heatmap", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	var hm map[string]interface{}
	decoder := json.NewDecoder(rw.Body)
	err := decoder.Decode(&hm)
	assert.Nil(t, err, err)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, hm["minTimestamp"], 0)
	assert.Equal(t, hm["maxTimestamp"], 0)
	assert.Equal(t, hm["maxValue"], 0)
	assert.Equal(t, hm["map"].([]interface{})[heatMapHeight-1].([]interface{})[heatMapWidth-1], 0)
	storageMock.Mock.AssertExpectations(t)
}

func TestHistogramMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getHistogram",
		"snapshot", "source", "cpu").Return(map[string]float64{"0.0 - 1.0": 100.0}, nil)

	controller := newController(storageMock)

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/histo", nil)
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"0.0 - 1.0\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

// PerfDB Unit Tests

func removeStorage(pdb *perfDB) {
	os.RemoveAll(pdb.BaseDir)
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
	var storage Storage
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
	var storage Storage
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
}

func TestListSourcesPerfDb(t *testing.T) {
	var err error
	var storage Storage
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/snapshot/source",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	req, _ = http.NewRequest("GET", "/snapshot", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"source\"]", rw.Body.String())
}

func TestListMetricsPerfDb(t *testing.T) {
	var err error
	var storage Storage
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/snapshot/source",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	req, _ = http.NewRequest("GET", "/snapshot/source", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"cpu\"]", rw.Body.String())
}

func TestGetRawValuesPerfDb(t *testing.T) {
	var err error
	var storage Storage
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	controller := newController(storage)

	req, _ := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":1005}"))
	rw := httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	req, _ = http.NewRequest("POST", "/snapshot/source?ts=1411940889515410775",
		bytes.NewBufferString("{\"cpu\":75.11}"))
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	req, _ = http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw = httptest.NewRecorder()
	newRouter(controller).ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t,
		"{\"1411940889515410774\":1005,\"1411940889515410775\":75.11}",
		rw.Body.String())
}
