package main

import (
	"testing"
)

func TestTimestampParser(t *testing.T) {
	// nanoseconds
	tsNano := ParseTimestamp("1411534805453497432")
	if tsNano != 1411534805453497432 {
		t.Errorf("Invalid ns: %v != %v", tsNano, 1411534805453497432)
	}
	// microseconds
	tsNano = ParseTimestamp("1411534805453497")
	if tsNano != 1411534805453497000 {
		t.Errorf("Invalid us: %v != %v", tsNano, 1411534805453497000)
	}
	// milliseconds
	tsNano = ParseTimestamp("1411534805453")
	if tsNano != 1411534805453000000 {
		t.Errorf("Invalid ms: %v != %v", tsNano, 1411534805453000000)
	}
	// seconds
	tsNano = ParseTimestamp("1411534805")
	if tsNano != 1411534805000000000 {
		t.Errorf("Invalid ms: %v != %v", tsNano, 1411534805000000000)
	}
}
