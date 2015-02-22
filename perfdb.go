package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type perfDb struct {
	BaseDir string
}

func newPerfDb(BaseDir string) (*perfDb, error) {
	if err := os.MkdirAll(BaseDir, 0755); err != nil {
		logger.Critical("Failed to initalize datastore: %s", err)
		return nil, err
	}
	return &perfDb{BaseDir}, nil
}

func (pdb *perfDb) listDatabases() ([]string, error) {
	files, err := ioutil.ReadDir(pdb.BaseDir)
	if err != nil {
		return nil, err
	}
	databases := []string{}
	for _, f := range files {
		databases = append(databases, f.Name())
	}
	return databases, nil
}

func (pdb *perfDb) listSources(dbname string) ([]string, error) {
	dstDir := filepath.Join(pdb.BaseDir, dbname)
	files, err := ioutil.ReadDir(dstDir)
	if err != nil {
		return nil, err
	}
	collections := []string{}
	for _, f := range files {
		collections = append(collections, f.Name())
	}
	return collections, nil
}

func (pdb *perfDb) listMetrics(dbname, collection string) ([]string, error) {
	dstDir := filepath.Join(pdb.BaseDir, dbname, collection)
	files, err := ioutil.ReadDir(dstDir)
	if err != nil {
		return nil, err
	}
	metrics := []string{}
	for _, f := range files {
		metrics = append(metrics, f.Name())
	}
	return metrics, nil
}

func (pdb *perfDb) getRawValues(dbname, collection, metric string) (map[string]float64, error) {
	dstDir := filepath.Join(pdb.BaseDir, dbname, collection)
	if err := os.MkdirAll(dstDir, 0775); err != nil {
		return nil, err
	}

	dstFile := filepath.Join(dstDir, metric)

	file, err := os.Open(dstFile)
	if err != nil {
		logger.Critical(err)
		return nil, err
	}
	defer file.Close()

	values := map[string]float64{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sample := strings.Fields(scanner.Text())
		var value float64
		if value, err = strconv.ParseFloat(sample[1], 64); err != nil {
			logger.Critical(err)
			return nil, err
		}
		values[sample[0]] = value
	}

	if err := scanner.Err(); err != nil {
		logger.Critical(err)
		return nil, err
	}

	return values, nil
}

func (pdb *perfDb) addSample(dbname, collection string, sample map[string]interface{}) error {
	dstDir := filepath.Join(pdb.BaseDir, dbname, collection)
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

	if _, err := fmt.Fprintf(file, "%s %025.12f\n", sample["ts"], sample["v"]); err != nil {
		logger.Critical(err)
		return err
	}
	return nil
}

func (pdb *perfDb) getSummary(dbname, collection, metric string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (pdb *perfDb) getHeatMap(dbname, collection, metric string) (*heatMap, error) {
	return newHeatMap(), nil
}

func (pdb *perfDb) getHistogram(dbname, collection, metric string) (map[string]float64, error) {
	return map[string]float64{}, nil
}
