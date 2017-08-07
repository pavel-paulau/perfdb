package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampParser(t *testing.T) {
	timestamp := parseTimestamp("1411534805453497432")
	assert.Equal(t, timestamp, int64(1411534805453), "Invalid ns")

	timestamp = parseTimestamp("1411534805453497")
	assert.Equal(t, timestamp, int64(1411534805453), "Invalid us")

	timestamp = parseTimestamp("1411534805453")
	assert.Equal(t, timestamp, int64(1411534805453), "Invalid ms")

	timestamp = parseTimestamp("1411534805")
	assert.Equal(t, timestamp, int64(1411534805000), "Invalid s")
}

func TestBadTimestampParser(t *testing.T) {
	timeNow := time.Now().UnixNano() / 1e6

	timestamp := parseTimestamp("1411534805.453")

	if timestamp < timeNow || timestamp > (timeNow+1e3) {
		t.Fatalf("Bad (not current) time: %v, expected ~%v", timestamp, timeNow)
	}
}

func TestSmallTimestampParser(t *testing.T) {
	timeNow := time.Now().UnixNano() / 1e6

	timestamp := parseTimestamp("123456")
	if timestamp < timeNow || timestamp > (timeNow+1e3) {
		t.Fatalf("Bad (not current) time: %v, expected ~%v", timestamp, timeNow)
	}
}
