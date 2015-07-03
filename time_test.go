package main
 
import (
	"testing"
 
	"github.com/stretchr/testify/assert"
)
 
func TestTimestampParser(t *testing.T) {
	tsNano := parseTimestamp("1411534805453497432")
	assert.Equal(t, tsNano, int64(1411534805453497432), "Invalid ns")
 
	tsNano = parseTimestamp("1411534805453497")
	assert.Equal(t, tsNano, int64(1411534805453497000), "Invalid us")
 
	tsNano = parseTimestamp("1411534805453")
	assert.Equal(t, tsNano, int64(1411534805453000000), "Invalid ms")
 
	tsNano = parseTimestamp("1411534805")
	assert.Equal(t, tsNano, int64(1411534805000000000), "Invalid s")
}
