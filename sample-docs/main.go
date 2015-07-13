package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
    "os"
	"runtime"
	"sync"

    "github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
)

type perfDbClient struct {
	client *http.Client
	uri    string
}

func newPerfDbClient(host, snapshot, source string) *perfDbClient {
	return &perfDbClient{
		client: &http.Client{},
		uri:    fmt.Sprintf("http://%s/%s/%s", host, snapshot, source),
	}
}

func (c *perfDbClient) store(sample map[string]uint64) error {
	b, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	j := bytes.NewReader(b)

	req, err := http.NewRequest("POST", c.uri, j)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

    return nil
}

const (
	bufferSize = 1e3
	totalNumSamples = 1e5
)

func randFloat64(numSamples int) <-chan uint64 {
	values := make(chan uint64, bufferSize)

	go func() {
		defer close(values)

		src := rand.NewSource(0)
		r := rand.New(src)
		zipf := rand.NewZipf(r, 5.0, 20.0, 100)

		for i := 0; i < numSamples; i++ {
			values <- zipf.Uint64()
		}
	}()
	return values
}

func runWorkload(numSamples int, client *perfDbClient, wg *sync.WaitGroup) {
	for value := range randFloat64(numSamples) {
		sample := map[string]uint64{"metric": value}
		if err := client.store(sample); err != nil {
            logger.Critical(err)
            break
        }
	}
	wg.Done()
}

var logger *golog.Logger

func init() {
    logger = golog.New(os.Stdout, log.Info)
}

func main() {
	numWorkers := 2 * runtime.NumCPU()
    numSamples := totalNumSamples / numWorkers
	client := newPerfDbClient("127.0.0.1:8080", "snapshot", "source")

    logger.Info("Loading sample data set. It will take a while...")

	wg := sync.WaitGroup{}
	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go runWorkload(numSamples, client, &wg)
	}
	wg.Wait()

    logger.Info("Done loading data")
}
