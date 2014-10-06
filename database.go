package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type storageHandler interface {
	listDatabases() ([]string, error)
	listCollections(dbname string) ([]string, error)
	listMetrics(dbname, collection string) ([]string, error)
	insertSample(dbname, collection string, sample map[string]interface{}) error
	findValues(dbname, collection, metric string) (map[string]float64, error)
	aggregate(dbname, collection, metric string) (map[string]interface{}, error)
}

type mongoHandler struct {
	Session *mgo.Session
}

func newMongoHandler() (*mongoHandler, error) {
	dialInfo := &mgo.DialInfo{
		Addrs:   []string{"127.0.0.1"},
		Timeout: 30 * time.Second,
	}

	logger.Info("Connecting to database...")
	if session, err := mgo.DialWithInfo(dialInfo); err != nil {
		logger.Criticalf("Failed to connect to database: %s", err)
		return nil, err
	} else {
		logger.Info("Connection established.")
		session.SetMode(mgo.Monotonic, true)
		return &mongoHandler{session}, nil
	}
}

var dbPrefix = "perf"

func (mongo *mongoHandler) listDatabases() ([]string, error) {
	allDbs, err := mongo.Session.DatabaseNames()
	if err != nil {
		logger.Critical(err)
		return nil, err
	}

	dbs := []string{}
	for _, db := range allDbs {
		if strings.HasPrefix(db, dbPrefix) {
			dbs = append(dbs, strings.Replace(db, dbPrefix, "", 1))
		}
	}
	return dbs, nil
}

func (mongo *mongoHandler) listCollections(dbname string) ([]string, error) {
	session := mongo.Session.New()
	defer session.Close()
	_db := session.DB(dbPrefix + dbname)

	allCollections, err := _db.CollectionNames()
	if err != nil {
		logger.Critical(err)
		return []string{}, err
	}

	collections := []string{}
	for _, collection := range allCollections {
		if collection != "system.indexes" {
			collections = append(collections, collection)
		}
	}
	return collections, err
}

func (mongo *mongoHandler) listMetrics(dbname, collection string) ([]string, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	var metrics []string
	if err := _collection.Find(bson.M{}).Distinct("m", &metrics); err != nil {
		logger.Critical(err)
		return []string{}, err
	}
	return metrics, nil
}

func (mongo *mongoHandler) findValues(dbname, collection, metric string) (map[string]float64, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	var docs []map[string]interface{}
	if err := _collection.Find(bson.M{"m": metric}).Sort("ts").All(&docs); err != nil {
		logger.Critical(err)
		return map[string]float64{}, err
	}
	values := map[string]float64{}
	for _, doc := range docs {
		values[doc["ts"].(string)] = doc["v"].(float64)
	}
	return values, nil
}

func (mongo *mongoHandler) insertSample(dbname, collection string, sample map[string]interface{}) error {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	if err := _collection.Insert(sample); err != nil {
		logger.Critical(err)
		return err
	}
	logger.Infof("Successfully added new sample to %s.%s", dbname, collection)

	for _, key := range []string{"m", "ts", "v"} {
		if err := _collection.EnsureIndexKey(key); err != nil {
			logger.Critical(err)
			return err
		}
	}
	return nil
}

func calcPercentile(data []float64, p float64) float64 {
	sort.Float64s(data)

	k := float64(len(data)-1) * p
	f := math.Floor(k)
	c := math.Ceil(k)
	if f == c {
		return data[int(k)]
	} else {
		return data[int(f)]*(c-k) + data[int(c)]*(k-f)
	}
}

var AggLimit = 10000

func (mongo *mongoHandler) aggregate(dbname, collection, metric string) (map[string]interface{}, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	pipe := _collection.Pipe(
		[]bson.M{
			{
				"$match": bson.M{
					"m": metric,
				},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"metric": "$m",
					},
					"avg": bson.M{"$avg": "$v"},
					"min": bson.M{"$min": "$v"},
					"max": bson.M{"$max": "$v"},
				},
			},
		},
	)
	summaries := []map[string]interface{}{}
	if err := pipe.All(&summaries); err != nil {
		logger.Critical(err)
		return map[string]interface{}{}, err
	}
	summary := summaries[0]
	delete(summary, "_id")

	var count int
	var err error
	if count, err = _collection.Find(bson.M{"m": metric}).Count(); err != nil {
		logger.Critical(err)
		return map[string]interface{}{}, err
	}

	if count < AggLimit {
		// Don't perform in-memory aggregation if limit exceeded
		var docs []map[string]interface{}
		if err := _collection.Find(bson.M{"m": metric}).Select(bson.M{"v": 1}).All(&docs); err != nil {
			logger.Critical(err)
			return map[string]interface{}{}, err
		}
		values := []float64{}
		for _, doc := range docs {
			values = append(values, doc["v"].(float64))
		}
		for _, percentile := range []float64{0.5, 0.8, 0.9, 0.95, 0.99} {
			p := fmt.Sprintf("p%v", percentile*100)
			summary[p] = calcPercentile(values, percentile)
		}
	} else {
		// Calculate percentiles using index-based sorting at database level
		var result []map[string]interface{}
		for _, percentile := range []float64{0.5, 0.8, 0.9, 0.95, 0.99} {
			skip := int(float64(count)*percentile) - 1
			if err := _collection.Find(bson.M{"m": metric}).Sort("v").Skip(skip).Limit(1).All(&result); err != nil {
				logger.Critical(err)
				return map[string]interface{}{}, err
			}
			p := fmt.Sprintf("p%v", percentile*100)
			summary[p] = result[0]["v"].(float64)
		}
	}

	return summary, nil
}
