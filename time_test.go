package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampParser(t *testing.T) {
	timestamp := parseTimestamp("1411534805453497432")
	assert.Equal(t, timestamp, int64(1411534805453497432), "Invalid ns")

	timestamp = parseTimestamp("1411534805453497")
	assert.Equal(t, timestamp, int64(1411534805453497000), "Invalid us")

	timestamp = parseTimestamp("1411534805453")
	assert.Equal(t, timestamp, int64(1411534805453000000), "Invalid ms")

	timestamp = parseTimestamp("1411534805")
	assert.Equal(t, timestamp, int64(1411534805000000000), "Invalid s")
}

func TestBadTimestampParser(t *testing.T) {
	timeNow := time.Now().UnixNano()

	timestamp := parseTimestamp("1411534805.453")

	if timestamp < timeNow || timestamp > (timeNow+1e9) {
		t.Fatalf("Bad (not current) time: %v, expected ~%v", timestamp, timeNow)
	}
}

func TestSmallTimestampParser(t *testing.T) {
	timeNow := time.Now().UnixNano()

	timestamp := parseTimestamp("123456")
	if timestamp < timeNow || timestamp > (timeNow+1e9) {
		t.Fatalf("Bad (not current) time: %v, expected ~%v", timestamp, timeNow)
	}
}
