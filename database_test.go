package main

import (
	"github.com/stretchr/testify/mock"
)

type mongoMock struct {
	mock.Mock
}

func (m *mongoMock) listDatabases() ([]string, error) {
	args := m.Mock.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *mongoMock) listCollections(dbname string) ([]string, error) {
	args := m.Mock.Called(dbname)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mongoMock) listMetrics(dbname, collection string) ([]string, error) {
	args := m.Mock.Called(dbname, collection)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mongoMock) findValues(dbname, collection, metric string) (map[string]float64, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *mongoMock) insertSample(dbname, collection string, sample map[string]interface{}) error {
	args := m.Mock.Called(dbname, collection, sample)
	return args.Error(0)
}

func (m *mongoMock) aggregate(dbname, collection, metric string) (map[string]interface{}, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mongoMock) getHeatMap(dbname, collection, metric string) (*heatMap, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(*heatMap), args.Error(1)
}
