package main

import (
	"github.com/stretchr/testify/mock"
)

type MongoMock struct {
	mock.Mock
}

func (m *MongoMock) ListDatabases() ([]string, error) {
	args := m.Mock.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MongoMock) ListCollections(dbname string) ([]string, error) {
	args := m.Mock.Called(dbname)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MongoMock) ListMetrics(dbname, collection string) ([]string, error) {
	args := m.Mock.Called(dbname, collection)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MongoMock) FindValues(dbname, collection, metric string) (map[string]float64, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MongoMock) InsertSample(dbname, collection string, sample map[string]interface{}) error {
	args := m.Mock.Called(dbname, collection, sample)
	return args.Error(0)
}

func (m *MongoMock) Aggregate(dbname, collection, metric string) (map[string]interface{}, error) {
	args := m.Mock.Called(dbname, collection, metric)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
