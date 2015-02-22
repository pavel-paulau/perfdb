package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type perfDb struct {
	Basedir string
}

func newPerfDb(basedir string) (*perfDb, error) {
	if err := os.MkdirAll(basedir, 0755); err != nil {
		logger.Critical("Failed to initalize datastore: %s", err)
		return nil, err
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
	dstDir := filepath.Join(pdb.Basedir, dbname, collection)
	if err := os.MkdirAll(dstDir, 0775); err != nil {
		return err
	}

	dstFile := filepath.Join(dstDir, sample["m"].(string))

	file, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logger.Critical(err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	value := strconv.FormatFloat(sample["v"].(float64), 'f', 12, 64)
	if err := writer.Write([]string{sample["ts"].(string), value}); err != nil {
		logger.Critical(err)
		return err
	}
	writer.Flush()

	return writer.Error()
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
