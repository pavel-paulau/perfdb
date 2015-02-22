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

func TestInterfacePerfDb(t *testing.T) {
	var err error
	var storage interface{}
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, nil, storage.(storageHandler))
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

func TestAddSamplePerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	if err = storage.addSample("testdb", "testcoll", sample); err != nil {
		t.Fatal(err)
	}
}

func TestListSourcesPerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	err = storage.addSample("testdb", "testcoll", sample)
	if err = storage.addSample("testdb", "testcoll", sample); err != nil {
		t.Fatal(err)
	}

	var collections []string
	if collections, err = storage.listSources("testdb"); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{"testcoll"}, collections)
}

func TestListMetricsPerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	if err = storage.addSample("testdb", "testcoll", sample); err != nil {
		t.Fatal(err)
	}

	var metrics []string
	if metrics, err = storage.listMetrics("testdb", "testcoll"); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{"cpu"}, metrics)
}

func TestGetRawValuePerfDb(t *testing.T) {
	var err error
	var storage *perfDb
	if storage, err = newTmpStorage(); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": float64(1005)}
	if err = storage.addSample("testdb", "testcoll", sample); err != nil {
		t.Fatal(err)
	}
	sample = map[string]interface{}{"ts": "1411940889515410775", "m": "cpu", "v": 75.11}
	if err = storage.addSample("testdb", "testcoll", sample); err != nil {
		t.Fatal(err)
	}

	var values map[string]float64
	if values, err = storage.getRawValues("testdb", "testcoll", "cpu"); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, map[string]float64{"1411940889515410774": 1005, "1411940889515410775": 75.11}, values)
}
