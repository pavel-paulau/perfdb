package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/pmylund/go-cache"
	"time"
)

const (
	dataFileExt = ".data"
)

type Sample struct {
	ts int64
	v  float64
}

type perfDB struct {
	baseDir string
	mu      sync.Mutex
}

var timestampCache *cache.Cache

func newPerfDB(baseDir string) (*perfDB, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		logger.Critical("Failed to initialize datastore: %s", err)
		return nil, err
	}
	timestampCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	return &perfDB{baseDir, sync.Mutex{}}, nil
}

func (pdb *perfDB) getDirPath(dbname string) string {
	return filepath.Join(pdb.baseDir, dbname)
}

func (pdb *perfDB) getFilePath(dbname, metric string) string {
	dataDir := pdb.getDirPath(dbname)

	return filepath.Join(dataDir, metric+dataFileExt)
}

func (pdb *perfDB) isExist(dbname string) (bool, error) {
	dataDir := pdb.getDirPath(dbname)

	_, err := os.Stat(dataDir)
	if err == nil {
		return true, nil
	}
	return false, err
}

func parseRecord(record string) (Sample, error) {
	fields := strings.Fields(record)

	ts, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return Sample{}, err
	}

	value, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return Sample{}, err
	}
	return Sample{ts, value}, nil
}

func readTimestamp(dataFile string) (int64, error) {
	if cachedData, found := timestampCache.Get(dataFile); found {
		return cachedData.(int64), nil
	}

	record, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(string(record), 10, 64)
}

func storeTimestamp(dataFile string, timestamp int64) error {
	file, err := os.OpenFile(dataFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = fmt.Fprint(file, timestamp); err != nil {
		return err
	}
	timestampCache.Set(dataFile, timestamp, time.Minute)
	return nil
}

func storeSample(dataFile string, sample Sample) error {
	file, err := os.OpenFile(dataFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d %v\n", sample.ts, sample.v)
	return err
}

func initStore(dataFile string, sample Sample) error {
	if err := storeTimestamp(dataFile+".1", sample.ts); err != nil {
		return err
	}
	if err := storeTimestamp(dataFile+".n", sample.ts); err != nil {
		return err
	}

	sample.ts = 0
	return storeSample(dataFile, sample)
}

func appendSample(dataFile string, sample Sample) error {
	ts, err := readTimestamp(dataFile + ".n")
	if err != nil {
		return err
	}
	if err := storeTimestamp(dataFile+".n", sample.ts); err != nil {
		return err
	}
	sample.ts -= ts
	return storeSample(dataFile, sample)
}

func (pdb *perfDB) addSample(dbname, metric string, sample Sample) error {
	dataDir := pdb.getDirPath(dbname)
	if err := os.MkdirAll(dataDir, 0775); err != nil {
		return err
	}
	dataFile := pdb.getFilePath(dbname, metric)

	pdb.mu.Lock()
	defer pdb.mu.Unlock()

	_, err := os.Stat(dataFile + ".n")
	if err == nil { // Append delta
		return appendSample(dataFile, sample)
	} else if os.IsNotExist(err) {
		return initStore(dataFile, sample)
	}
	return err
}

func (pdb *perfDB) listDatabases() ([]string, error) {
	files, err := ioutil.ReadDir(pdb.baseDir)
	if err != nil {
		return nil, err
	}

	databases := []string{}
	for _, f := range files {
		databases = append(databases, f.Name())
	}
	return databases, nil
}

func (pdb *perfDB) listMetrics(dbname string) ([]string, error) {
	dataDir := pdb.getDirPath(dbname)

	matches, err := filepath.Glob(dataDir + "/*.data")
	if err != nil {
		return nil, err
	}

	metrics := []string{}
	for _, match := range matches {
		name := strings.TrimPrefix(match, dataDir+"/")
		name = strings.TrimRight(name, dataFileExt)
		metrics = append(metrics, name)
	}
	return metrics, nil
}

const bufferSize = 1000

func readDeltas(fileName string, done <-chan struct{}) (<-chan string, <-chan error) {
	samples := make(chan string, bufferSize)
	errc := make(chan error, 1)

	go func() {
		defer close(samples)
		defer close(errc)

		file, err := os.Open(fileName)
		if err != nil {
			errc <- err
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			select {
			case samples <- scanner.Text():
			case <-done:
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errc <- err
		}
	}()
	return samples, errc
}

func parseSamples(records <-chan string, ts int64) (<-chan Sample, <-chan error) {
	samples := make(chan Sample, bufferSize)
	errc := make(chan error, 1)

	go func() {
		defer close(samples)
		defer close(errc)

		for record := range records {
			sample, err := parseRecord(record)
			if err != nil {
				errc <- err
			} else {
				sample.ts += ts
				ts = sample.ts
				samples <- sample
			}
		}
	}()
	return samples, errc
}

func mergeErrors(errcs ...<-chan error) error {
	for _, errc := range errcs {
		if err := <-errc; err != nil {
			return err
		}
	}
	return nil
}

func (pdb *perfDB) getRawValues(dbname, metric string) ([][]interface{}, error) {
	dataFile := pdb.getFilePath(dbname, metric)

	first, err := readTimestamp(dataFile + ".1")
	if err != nil {
		return nil, err
	}

	done := make(chan struct{}, 1)
	rawSamples, rawErrors := readDeltas(dataFile, done)
	parsedSamples, parsedErrors := parseSamples(rawSamples, first)

	values := [][]interface{}{}
	for sample := range parsedSamples {
		values = append(values, []interface{}{sample.ts, sample.v})
	}

	done <- struct{}{}
	if err := mergeErrors(rawErrors, parsedErrors); err != nil {
		return nil, err
	}
	return values, nil
}

func (pdb *perfDB) getSummary(dbname, metric string) (map[string]interface{}, error) {
	var summary map[string]interface{}

	dataFile := pdb.getFilePath(dbname, metric)

	done := make(chan struct{}, 1)
	defer close(done)

	first, err := readTimestamp(dataFile + ".1")
	if err != nil {
		return nil, err
	}

	rawSamples, rawErrors := readDeltas(dataFile, done)
	parsedSamples, parsedErrors := parseSamples(rawSamples, first)

	values := []float64{}
	sum := 0.0
	for sample := range parsedSamples {
		sum += sample.v
		values = append(values, sample.v)
	}

	done <- struct{}{}
	if err := mergeErrors(rawErrors, parsedErrors); err != nil {
		return nil, err
	}

	count := len(values)
	sort.Float64s(values)

	summary = map[string]interface{}{
		"max":   values[count-1],
		"min":   values[0],
		"count": count,
		"avg":   sum / float64(count),
	}

	for _, percentile := range []float64{0.5, 0.8, 0.9, 0.95, 0.99, 0.999} {
		var pIdx int
		if count > 1 {
			pIdx = int(float64(count)*percentile) - 1
		}
		p := fmt.Sprintf("p%v", percentile*100)
		summary[p] = values[pIdx]
	}

	return summary, nil
}

func (pdb *perfDB) getHeatMap(dbname, metric string) (*heatMap, error) {
	dataFile := pdb.getFilePath(dbname, metric)

	hm := newHeatMap()
	hm.MinTS = int64(^uint64(0) >> 1)

	done := make(chan struct{}, 1)
	defer close(done)

	first, err := readTimestamp(dataFile + ".1")
	if err != nil {
		return nil, err
	}

	rawSamples, rawErrors := readDeltas(dataFile, done)
	parsedSamples, parsedErrors := parseSamples(rawSamples, first)

	samples := []Sample{}
	for sample := range parsedSamples {
		hm.MaxValue = math.Max(hm.MaxValue, sample.v)
		if sample.ts < hm.MinTS {
			hm.MinTS = sample.ts
		} else if sample.ts > hm.MaxTS {
			hm.MaxTS = sample.ts
		}
		samples = append(samples, sample)
	}

	done <- struct{}{}
	if err := mergeErrors(rawErrors, parsedErrors); err != nil {
		return nil, err
	}

	for _, sample := range samples {
		x := math.Floor(heatMapWidth * float64(sample.ts-hm.MinTS) / float64(hm.MaxTS-hm.MinTS))
		y := math.Floor(heatMapHeight * sample.v / hm.MaxValue)
		if x == heatMapWidth {
			x--
		}
		if y == heatMapHeight {
			y--
		}
		hm.Map[int(y)][int(x)]++
		if hm.Map[int(y)][int(x)] > hm.maxDensity {
			hm.maxDensity = hm.Map[int(y)][int(x)]
		}
	}
	return hm, nil
}
