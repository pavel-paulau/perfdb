package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func removeStorage(pdb *perfDb) {
	os.RemoveAll(pdb.BaseDir)
}

func newTmpStorage() (*perfDb, error) {
	var tmpDir string
	var err error

	if tmpDir, err = ioutil.TempDir("", ""); err != nil {
		return nil, err
	}

	var storage *perfDb
	if storage, err = newPerfDb(tmpDir); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(storage, removeStorage)
	return storage, nil
}

func TestListDatabasesPerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	var databases []string
	if databases, err = storage.listDatabases(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{}, databases)
}

func TestAddSampleMongoPerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	err = storage.addSample("testdb", "testcoll", sample)
	assert.Nil(t, err)

	sample = map[string]interface{}{"ts": "1411940889515410775", "m": "cpu", "v": 75.11}
	err = storage.addSample("testdb", "testcoll", sample)
	assert.Nil(t, err)

}

func TestListSourcesPerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	err = storage.addSample("testdb", "testcoll", sample)
	assert.Nil(t, err)

	collections, err := storage.listSources("testdb")
	assert.Nil(t, err)
	assert.Equal(t, []string{"testcoll"}, collections)
}

func TestListMetricsPerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	err = storage.addSample("testdb", "testcoll", sample)
	assert.Nil(t, err)

	metrics, err := storage.listMetrics("testdb", "testcoll")
	assert.Nil(t, err)
	assert.Equal(t, []string{"cpu"}, metrics)
}