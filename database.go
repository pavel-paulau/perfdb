package main

import (
	"log"
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
}
