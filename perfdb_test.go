package main

import (
	"io/ioutil"
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
}
