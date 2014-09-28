package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListDatabases(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("ListDatabases").Return([]string{"snapshot"}, nil)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"snapshot\"]", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestListCollections(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("ListCollections",
		"snapshot").Return([]string{"source"}, nil)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"source\"]", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

var DbError = errors.New("Database error")

func TestListCollectionsError(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("ListCollections",
		"snapshot").Return([]string{}, DbError)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestListMetrics(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("ListMetrics",
		"snapshot", "source").Return([]string{"cpu_user", "cpu_sys"}, nil)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"cpu_user\",\"cpu_sys\"]", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestListMetricsError(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("ListMetrics",
		"snapshot", "source").Return([]string{}, DbError)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestFindValues(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("FindValues",
		"snapshot", "source", "cpu").Return(map[string]float64{"1411534805453497432": 100}, nil)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"1411534805453497432\":100}", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestFinaValuesError(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("FindValues",
		"snapshot", "source", "cpu").Return(map[string]float64{}, DbError)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestInsertSample(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On("InsertSample",
		"snapshot", "source", map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}).Return(nil)
	Storage = StorageMock

	req, _ := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestInsertSampleError(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On(
		"InsertSample",
		"snapshot",
		"source",
		map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0},
	).Return(DbError)
	Storage = StorageMock

	req, err := http.NewRequest("POST", "/snapshot/source?ts=1411940889515410774",
		bytes.NewBufferString("{\"cpu\":99.0}"))
	if err != nil {
		log.Fatal(err)
	}
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)
	time.Sleep(10 * time.Millisecond) // Goroutine

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestInsertBadSample(t *testing.T) {
	req, _ := http.NewRequest("POST", "/snapshot/source",
		bytes.NewBufferString(""))
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 400, rw.Code)
	assert.Equal(t, "Cannot decode sample: EOF\n", rw.Body.String())
}

func TestSummary(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On(
		"Aggregate",
		"snapshot",
		"source",
		"cpu",
	).Return(map[string]interface{}{"mean": 100}, nil)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"mean\":100}", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}

func TestSummaryError(t *testing.T) {
	StorageMock := new(MongoMock)
	StorageMock.Mock.On(
		"Aggregate",
		"snapshot",
		"source",
		"cpu",
	).Return(map[string]interface{}{}, DbError)
	Storage = StorageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	NewRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	StorageMock.Mock.AssertExpectations(t)
}
