package main

type Sample struct {
	ts string
	v  float64
}

type storageHandler interface {
	listDatabases() ([]string, error)
	listSources(dbname string) ([]string, error)
	listMetrics(dbname, collection string) ([]string, error)
	addSample(dbname, collection, metric string, sample Sample) error
	getRawValues(dbname, collection, metric string) (map[string]float64, error)
	getSummary(dbname, collection, metric string) (map[string]float64, error)
	getHeatMap(dbname, collection, metric string) (*heatMap, error)
	getHistogram(dbname, collection, metric string) (map[string]float64, error)
}
