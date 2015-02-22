package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/tebeka/nrsc"
	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
)

var (
	logger      *golog.Logger
	db, address *string
	timeout     *time.Duration
	storage     storageHandler
)

type storageHandler interface {
	listDatabases() ([]string, error)
	listSources(dbname string) ([]string, error)
	listMetrics(dbname, collection string) ([]string, error)
	addSample(dbname, collection string, sample map[string]interface{}) error
	getRawValues(dbname, collection, metric string) (map[string]float64, error)
	getSummary(dbname, collection, metric string) (map[string]interface{}, error)
	getHeatMap(dbname, collection, metric string) (*heatMap, error)
	getHistogram(dbname, collection, metric string) (map[string]float64, error)
}

func requestLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Infof("%s %s", r.Method, r.URL)
		handler.ServeHTTP(rw, r)
	})
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	address = flag.String("address", "127.0.0.1:8080", "serve requests to this host[:port]")
	db = flag.String("db", "127.0.0.1:27017", "comma separated database host[:port] addresses")
	timeout = flag.Duration("timeout", 30*time.Second, "request timeout")
	flag.Parse()

	logger = golog.New(os.Stdout, log.Info)
}

func main() {
	// Database handler
	var err error
	if storage, err = newMongoDB(strings.Split(*db, ","), *timeout); err != nil {
		os.Exit(1)
	}

	// Static assets
	nrsc.Handle("/static/")

	// RESTful API
	http.Handle("/", newRouter())

	// Banner and launcher
	banner := fmt.Sprintf("\n\t:-:-: perfkeeper :-:-:\t\t\tserving http://%s/\n", *address)
	fmt.Println(banner)
	logger.Critical(http.ListenAndServe(*address, requestLog(http.DefaultServeMux)))
}
