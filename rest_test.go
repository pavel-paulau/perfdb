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

func TestListCollections(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listCollections", "snapshot").Return([]string{"source"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"source\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

var ErrTest = errors.New("fake test error")

func TestListCollectionsError(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listCollections", "snapshot").Return([]string{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetrics(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listMetrics", "snapshot", "source").Return([]string{"cpu_user", "cpu_sys"}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "[\"cpu_user\",\"cpu_sys\"]", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestListMetricsError(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("listMetrics", "snapshot", "source").Return([]string{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestFindValues(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("findValues",
		"snapshot", "source", "cpu").Return(map[string]float64{"1411534805453497432": 100}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"1411534805453497432\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestFinaValuesError(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("findValues",
		"snapshot", "source", "cpu").Return(map[string]float64{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestInsertSample(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("insertSample",
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

func TestInsertSampleError(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("insertSample",
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

func TestInsertBadSample(t *testing.T) {
	req, _ := http.NewRequest("POST", "/snapshot/source",
		bytes.NewBufferString(""))
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 400, rw.Code)
	assert.Equal(t, "Cannot decode sample: EOF\n", rw.Body.String())
}

func TestSummary(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("aggregate",
		"snapshot", "source", "cpu").Return(map[string]interface{}{"mean": 100}, nil)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "{\"mean\":100}", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}

func TestSummaryError(t *testing.T) {
	storageMock := new(mongoMock)
	storageMock.Mock.On("aggregate", "snapshot", "source", "cpu").Return(
		map[string]interface{}{}, ErrTest)
	storage = storageMock

	req, _ := http.NewRequest("GET", "/snapshot/source/cpu/summary", nil)
	rw := httptest.NewRecorder()
	newRouter().ServeHTTP(rw, req)

	assert.Equal(t, 500, rw.Code)
	assert.Equal(t, "", rw.Body.String())
	storageMock.Mock.AssertExpectations(t)
}
