package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListDatabasesMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"snapshot\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListSourcesMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listSources", "snapshot").Return([]string{"source"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"source\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

var ErrTest = errors.New("fake test error")

func TestListSourcesErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listSources", "snapshot").Return([]string{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListSourcesWrongSnapshotMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshotx", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 404, rw.Code)
	assert.Equal(t, "{\"error\":\"not found\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listMetrics", "snapshot", "source").Return([]string{"cpu_user", "cpu_sys"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"cpu_user\",\"cpu_sys\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("listMetrics", "snapshot", "source").Return([]string{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsWrongSnapshotMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshotx/source", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 404, rw.Code)
	assert.Equal(t, "{\"error\":\"not found\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestFindValuesMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getRawValues",
		"snapshot", "source", "cpu").Return(map[string]float64{"1411534805453497432": 100}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"1411534805453497432\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestFinaValuesErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getRawValues",
		"snapshot", "source", "cpu").Return(map[string]float64{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestAddSampleMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("addSample",
		"snapshot", "source", map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0},
	).Return(nil)
	storage = storageMock

	req, _ := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestAddSampleErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("addSample",
		"snapshot", "source", map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0},
	).Return(ErrTest)
	storage = storageMock

	req, err := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	if err != nil {
		log.Fatal(err)
	}
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestAddBadSampleMongo(t *testing.T) {
	req, _ := http.NewRequest("POST", "/snapshot/source",
		bytes.NewBufferString(""))
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 400, rw.Code)
	assert.Equal(t, "{\"error\":\"EOF\"}", rw.Body.String())
}

func TestSummaryMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getSummary",
		"snapshot", "source", "cpu").Return(map[string]interface{}{"mean": 100}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"mean\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestSummaryErrorMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getSummary", "snapshot", "source", "cpu").Return(
		map[string]interface{}{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "{\"error\":\"fake test error\"}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestHeatMapMongo(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listDatabases").Return([]string{"snapshot"}, nil)
	storageMock.Mock.On("getHeatMap", "snapshot", "source", "cpu").Return(
		newHeatMap(), nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/heatmap", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

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
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/histo", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"0.0 - 1.0\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}
