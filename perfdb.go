package main

import (
	"io/ioutil"
	"os"
)

type perfDb struct {
	Basedir string
}

func newPerfDb(basedir string) (*perfDb, error) {
	if _, err := os.Stat(basedir); err != nil {
		if os.IsNotExist(err) {
			logger.Infof("Creating a new datastore '%s' ...", basedir)
			os.Mkdir(basedir, 755)
		}
	}
	return &perfDb{basedir}, nil
}

func (pdb *perfDb) listDatabases() ([]string, error) {
	files, err := ioutil.ReadDir(pdb.Basedir)
	if err != nil {
		return nil, err
	}
	databases := []string{}
	for _, f := range files {
		databases = append(databases, f.Name())
	}
	return databases, nil
}

func (pdb *perfDb) listCollections(dbname string) ([]string, error) {
	return []string{}, nil
}

func (pdb *perfDb) listMetrics(dbname, collection string) ([]string, error) {
	return []string{}, nil
}

func (pdb *perfDb) findValues(dbname, collection, metric string) (map[string]float64, error) {
	return map[string]float64{}, nil
}

func (pdb *perfDb) insertSample(dbname, collection string, sample map[string]interface{}) error {
	return nil
}

func (pdb *perfDb) aggregate(dbname, collection, metric string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (pdb *perfDb) getHeatMap(dbname, collection, metric string) (*heatMap, error) {
	return newHeatMap(), nil
}

func (pdb *perfDb) getHistogram(dbname, collection, metric string) (map[string]float64, error) {
	return map[string]float64{}, nil
}
