package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type mongoHandler struct {
	Session *mgo.Session
}

func newMongoHandler(addrs []string, timeout time.Duration) (*mongoHandler, error) {
	dialInfo := &mgo.DialInfo{
		Addrs:   addrs,
		Timeout: timeout,
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
	if err := mongo.Session.Ping(); err != nil {
		mongo.Session.Refresh()
	}
	allDbs, err := mongo.Session.DatabaseNames()
	if err != nil {
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

func (mongo *mongoHandler) listSources(dbname string) ([]string, error) {
	session := mongo.Session.New()
	defer session.Close()
	_db := session.DB(dbPrefix + dbname)

	allCollections, err := _db.CollectionNames()
	if err != nil {
		return nil, err
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
	if err := _collection.Find(bson.M{}).Sort("m").Distinct("m", &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (mongo *mongoHandler) getRawValues(dbname, collection, metric string) (map[string]float64, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	var docs []map[string]interface{}
	if err := _collection.Find(bson.M{"m": metric}).Sort("ts").All(&docs); err != nil {
		return nil, err
	}
	values := map[string]float64{}
	for _, doc := range docs {
		values[doc["ts"].(string)] = doc["v"].(float64)
	}
	return values, nil
}

func (mongo *mongoHandler) addSample(dbname, collection string, sample map[string]interface{}) error {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	if err := _collection.Insert(sample); err != nil {
		return err
	}
	logger.Infof("Successfully added new sample to %s.%s", dbname, collection)

	for _, key := range []string{"m", "ts", "v"} {
		if err := _collection.EnsureIndexKey(key); err != nil {
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
	}
	return data[int(f)]*(c-k) + data[int(c)]*(k-f)
}

const queryLimit = 10000

func (mongo *mongoHandler) getSummary(dbname, collection, metric string) (map[string]interface{}, error) {
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
					"_id":   nil,
					"avg":   bson.M{"$avg": "$v"},
					"min":   bson.M{"$min": "$v"},
					"max":   bson.M{"$max": "$v"},
					"count": bson.M{"$sum": 1},
				},
			},
			{
				"$project": bson.M{
					"_id":   0,
					"avg":   1,
					"min":   1,
					"max":   1,
					"count": 1,
				},
			},
		},
	)
	summaries := []map[string]interface{}{}
	if err := pipe.All(&summaries); err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return map[string]interface{}{}, nil
	}
	summary := summaries[0]

	count := summary["count"].(int)
	if count < queryLimit {
		// Don't perform in-memory aggregation if limit exceeded
		var docs []map[string]interface{}
		if err := _collection.Find(bson.M{"m": metric}).Select(bson.M{"v": 1}).All(&docs); err != nil {
			return nil, err
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
				return map[string]interface{}{}, err
			}
			p := fmt.Sprintf("p%v", percentile*100)
			summary[p] = result[0]["v"].(float64)
		}
	}

	return summary, nil
}

func (mongo *mongoHandler) getHeatMap(dbname, collection, metric string) (*heatMap, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	var doc map[string]interface{}
	hm := newHeatMap()

	// Min timestamp
	if err := _collection.Find(bson.M{"m": metric}).Sort("ts").One(&doc); err != nil {
		return nil, err
	}
	if tsInt, err := strconv.ParseInt(doc["ts"].(string), 10, 64); err != nil {
		return nil, err
	} else {
		hm.MinTS = tsInt
	}
	// Max timestamp
	if err := _collection.Find(bson.M{"m": metric}).Sort("-ts").One(&doc); err != nil {
		return nil, err
	}
	if tsInt, err := strconv.ParseInt(doc["ts"].(string), 10, 64); err != nil {
		return nil, err
	} else {
		hm.MaxTS = tsInt
	}
	// Max value
	if err := _collection.Find(bson.M{"m": metric}).Sort("-v").One(&doc); err != nil {
		return nil, err
	}
	hm.MaxValue = doc["v"].(float64)

	if hm.MaxTS == hm.MinTS || hm.MaxValue == 0 {
		return hm, nil
	}

	iter := _collection.Find(bson.M{"m": metric}).Sort("ts").Iter()
	for iter.Next(&doc) {
		tsInt, err := strconv.ParseInt(doc["ts"].(string), 10, 64)
		if err != nil {
			return nil, err
		}
		x := math.Floor(heatMapWidth * float64(tsInt-hm.MinTS) / float64(hm.MaxTS-hm.MinTS))
		y := math.Floor(heatMapHeight * doc["v"].(float64) / hm.MaxValue)
		if x == heatMapWidth {
			x--
		}
		if y == heatMapHeight {
			y--
		}
		hm.Map[int(y)][int(x)]++
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return hm, nil
}

const numBins = 6

func (mongo *mongoHandler) getHistogram(dbname, collection, metric string) (map[string]float64, error) {
	session := mongo.Session.New()
	defer session.Close()
	_collection := session.DB(dbPrefix + dbname).C(collection)

	count, err := _collection.Find(bson.M{"m": metric}).Count()
	if err != nil {
		return nil, err
	}
	if count <= 1 {
		return nil, errors.New("not enough data points")
	}
	skip := int(float64(count)*0.99) - 1

	var doc map[string]interface{}
	if err := _collection.Find(bson.M{"m": metric}).Sort("v").Skip(skip).One(&doc); err != nil {
		return nil, err
	}
	p99 := doc["v"].(float64)
	if err := _collection.Find(bson.M{"m": metric}).Sort("v").One(&doc); err != nil {
		return nil, err
	}
	minValue := doc["v"].(float64)
	if p99 == minValue {
		return nil, errors.New("dataset lacks variation")
	}

	delta := (p99 - minValue) / numBins
	histogram := map[string]float64{}
	for i := 0; i < numBins; i++ {
		lr := minValue + float64(i)*delta
		rr := lr + delta
		rname := fmt.Sprintf("%f - %f", lr, rr)
		count, err := _collection.Find(bson.M{"m": metric, "v": bson.M{"$gte": lr, "$lt": rr}}).Count()
		if err != nil {
			return nil, err
		}
		histogram[rname] = 100.0 * float64(count) / float64(skip)
	}

	return histogram, nil
}
