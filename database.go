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

type StorageHandler interface {
	ListDatabases() ([]string, error)
	ListCollections(dbname string) ([]string, error)
	ListMetrics(dbname, collection string) ([]string, error)
	InsertSample(dbname, collection string, sample map[string]interface{}) error
	FindValues(dbname, collection, metric string) (map[string]float64, error)
	Aggregate(dbname, collection, metric string) (map[string]interface{}, error)
}

type MongoHandler struct {
	Session *mgo.Session
}

func NewMongoHandler() (*MongoHandler, error) {
	dialInfo := &mgo.DialInfo{
		Addrs:   []string{"127.0.0.1"},
		Timeout: 30 * time.Second,
	}

	Logger.Info("Connecting to database...")
	if session, err := mgo.DialWithInfo(dialInfo); err != nil {
		Logger.Criticalf("Failed to connect to database: %s", err)
		return nil, err
	} else {
		Logger.Info("Connection established.")
		session.SetMode(mgo.Monotonic, true)
		return &MongoHandler{session}, nil
	}
}

var DBPREFIX = "perf"

func (mongo *MongoHandler) ListDatabases() ([]string, error) {
	all_dbs, err := mongo.Session.DatabaseNames()
	if err != nil {
		Logger.Critical(err)
		return nil, err
	}

	dbs := []string{}
	for _, db := range all_dbs {
		if strings.HasPrefix(db, DBPREFIX) {
			dbs = append(dbs, strings.Replace(db, DBPREFIX, "", 1))
		}
	}
	return dbs, nil
}

func (mongo *MongoHandler) ListCollections(dbname string) ([]string, error) {
	session := mongo.Session.New()
	defer session.Close()
	_db := session.DB(DBPREFIX + dbname)

	all_collections, err := _db.CollectionNames()
	if err != nil {
		Logger.Critical(err)
		return []string{}, err
	}

	collections := []string{}
	for _, collection := range all_collections {
		if collection != "system.indexes" {
			collections = append(collections, collection)
		}
	}
	return collections, err
}

func (mongo *MongoHandler) ListMetrics(dbname, collection string) ([]string, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(DBPREFIX + dbname).C(collection)

	var metrics []string
	if err := _collection.Find(bson.M{}).Distinct("m", &metrics); err != nil {
		Logger.Critical(err)
		return []string{}, err
	} else {
		return metrics, nil
	}
}

func (mongo *MongoHandler) FindValues(dbname, collection, metric string) (map[string]float64, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(DBPREFIX + dbname).C(collection)

	var docs []map[string]interface{}
	if err := _collection.Find(bson.M{"m": metric}).Sort("ts").All(&docs); err != nil {
		Logger.Critical(err)
		return map[string]float64{}, err
	} else {
		values := map[string]float64{}
		for _, doc := range docs {
			values[doc["ts"].(string)] = doc["v"].(float64)
		}
		return values, nil
	}
}

func (mongo *MongoHandler) InsertSample(dbname, collection string, sample map[string]interface{}) error {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(DBPREFIX + dbname).C(collection)

	if err := _collection.Insert(sample); err != nil {
		Logger.Critical(err)
		return err
	} else {
		Logger.Infof("Successfully added new sample to %s.%s", dbname, collection)
	}

	for _, key := range []string{"m", "ts"} {
		err := _collection.EnsureIndexKey(key)
		if err != nil {
			Logger.Critical(err)
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

func (mongo *MongoHandler) Aggregate(dbname, collection, metric string) (map[string]interface{}, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(DBPREFIX + dbname).C(collection)

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
		Logger.Critical(err)
		return map[string]interface{}{}, err
	}
	summary := summaries[0]
	delete(summary, "_id")

	var docs []map[string]interface{}
	if err := _collection.Find(bson.M{"m": metric}).Select(bson.M{"v": 1}).All(&docs); err != nil {
		Logger.Critical(err)
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

	return summary, nil
}
