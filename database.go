package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type MongoHandler struct {
	Session *mgo.Session
}

func (mongo *MongoHandler) Init() {
	dialInfo := &mgo.DialInfo{
		Addrs:   []string{"127.0.0.1"},
		Timeout: 10 * time.Minute,
	}

	var err error
	mongo.Session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}
	mongo.Session.SetMode(mgo.Monotonic, true)
}

func (mongo *MongoHandler) ListDatabases() []string {
	all_dbs, err := mongo.Session.DatabaseNames()
	if err != nil {
		log.Fatal(err)
	}

	dbs := []string{}
	for _, db := range all_dbs {
		if strings.HasPrefix(db, "perf") {
			dbs = append(dbs, strings.Replace(db, "perf", "", 1))
		}
	}
	return dbs
}

func (mongo *MongoHandler) ListCollections(db string) []string {
	session := mongo.Session.New()
	defer session.Close()
	_db := session.DB(db)

	all_collections, err := _db.CollectionNames()
	if err != nil {
		log.Fatal(err)
	}

	collections := []string{}
	for _, collection := range all_collections {
		if collection != "system.indexes" {
			collections = append(collections, collection)
		}
	}
	return collections
}

func (mongo *MongoHandler) ListMetrics(db, collection string) []string {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(db).C(collection)

	var metrics []string
	err := _collection.Find(bson.M{}).Distinct("m", &metrics)
	if err != nil {
		log.Fatal(err)
	}
	return metrics
}

func (mongo *MongoHandler) FindValues(db, collection, metric string) map[string]float64 {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(db).C(collection)

	var docs []map[string]interface{}
	err := _collection.Find(bson.M{"m": metric}).Sort("ts").All(&docs)
	if err != nil {
		log.Fatal(err)
	}

	values := map[string]float64{}
	for _, doc := range docs {
		values[doc["ts"].(string)] = doc["v"].(float64)
	}

	return values
}

func (mongo *MongoHandler) InsertSample(db, collection string, sample map[string]interface{}) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(db).C(collection)

	err := _collection.Insert(sample)
	if err != nil {
		log.Fatal(err)
	}

	err = _collection.EnsureIndexKey("m")
	if err != nil {
		log.Fatal(err)
	}
	err = _collection.EnsureIndexKey("ts")
	if err != nil {
		log.Fatal(err)
	}
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

func (mongo *MongoHandler) Aggregate(db, collection, metric string) map[string]interface{} {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(db).C(collection)

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
	err := pipe.All(&summaries)
	if err != nil {
		log.Fatal(err)
	}
	summary := summaries[0]
	delete(summary, "_id")

	var docs []map[string]interface{}
	err = _collection.Find(bson.M{"m": metric}).Select(bson.M{"v": 1}).All(&docs)
	if err != nil {
		log.Fatal(err)
	}
	values := []float64{}
	for _, doc := range docs {
		values = append(values, doc["v"].(float64))
	}
	for _, percentile := range []float64{0.8, 0.9, 0.95, 0.99} {
		p := fmt.Sprintf("p%v", percentile*100)
		summary[p] = calcPercentile(values, percentile)
	}

	return summary
}
