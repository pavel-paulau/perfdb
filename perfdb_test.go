package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPerfDb(t *testing.T) {
	var tmpDir string
	var err error

	if tmpDir, err = ioutil.TempDir("", ""); err != nil {
		t.Fatal(err)
	}

	var storage *perfDb
	if storage, err = newPerfDb(tmpDir); err != nil {
		t.Fatal(err)
	}

	var databases []string
	if databases, err = storage.listDatabases(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{}, databases)

	if storage, err = newPerfDb(tmpDir); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(tmpDir)
}

func TestInsertPerfDb(t *testing.T) {
	var tmpDir string
	var err error

	if tmpDir, err = ioutil.TempDir("", ""); err != nil {
		t.Fatal(err)
	}

	var storage *perfDb
	if storage, err = newPerfDb(tmpDir); err != nil {
		t.Fatal(err)
	}

	sample := map[string]interface{}{"ts": "1411940889515410774", "m": "cpu", "v": 99.0}
	err = storage.insertSample("testdb", "testcoll", sample)
	assert.Nil(t, err)

	sample = map[string]interface{}{"ts": "1411940889515410775", "m": "cpu", "v": 75.11}
	err = storage.insertSample("testdb", "testcoll", sample)
	assert.Nil(t, err)

	os.RemoveAll(tmpDir)
}
