package main

type Sample struct {
	ts int64
	v  float64
}

type Storage interface {
	listDatabases() ([]string, error)
	listSources(dbname string) ([]string, error)
	listMetrics(dbname, collection string) ([]string, error)
	addSample(dbname, collection, metric string, sample Sample) error
	getRawValues(dbname, collection, metric string) ([][]interface{}, error)
	getSummary(dbname, collection, metric string) (map[string]interface{}, error)
	getHeatMap(dbname, collection, metric string) (*heatMap, error)
	getHistogram(dbname, collection, metric string) (map[string]float64, error)
}
