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

func (m *mongoMock) listSources(dbname string) ([]string, error) {
	args := m.Mock.Called(dbname)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mongoMock) listMetrics(dbname, collection string) ([]string, error) {
	args := m.Mock.Called(dbname, collection)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mongoMock) getRawValues(dbname, collection, metric string) (map[string]float64, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *mongoMock) addSample(dbname, collection, metric string, sample Sample) error {
	args := m.Mock.Called(dbname, collection, metric, sample)
	return args.Error(0)
}

func (m *mongoMock) getSummary(dbname, collection, metric string) (map[string]float64, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *mongoMock) getHeatMap(dbname, collection, metric string) (*heatMap, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(*heatMap), args.Error(1)
}

func (m *mongoMock) getHistogram(dbname, collection, metric string) (map[string]float64, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]float64), args.Error(1)
}
