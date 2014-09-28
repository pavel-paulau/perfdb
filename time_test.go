package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimestampParser(t *testing.T) {
	tsNano := ParseTimestamp("1411534805453497432")
	assert.Equal(t, tsNano, 1411534805453497432, "Invalid ns")

	tsNano = ParseTimestamp("1411534805453497")
	assert.Equal(t, tsNano, 1411534805453497000, "Invalid us")

	tsNano = ParseTimestamp("1411534805453")
	assert.Equal(t, tsNano, 1411534805453000000, "Invalid ms")

	tsNano = ParseTimestamp("1411534805")
	assert.Equal(t, tsNano, 1411534805000000000, "Invalid s")
}
