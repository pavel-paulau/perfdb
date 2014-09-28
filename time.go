package main

import (
	"strconv"
	"time"
)

func parseTimestamp(ts string) int64 {
	if tsInt, err := strconv.ParseInt(ts, 10, 64); err == nil {
		switch {
		case tsInt > 1e18: // nanosecond timestamps
			return time.Unix(tsInt/1e9, tsInt%1e9).UnixNano()
		case tsInt > 1e15: // microseconds timestamps
			return time.Unix(tsInt/1e6, (tsInt*1e3)%1e9).UnixNano()
		case tsInt > 1e12: // millisecond timestamps
			return time.Unix(tsInt/1e3, (tsInt*1e6)%1e9).UnixNano()
		case tsInt > 1e9: // second timestamps
			return time.Unix(tsInt, 0).UnixNano()
		}
	} else {
		logger.Warningf("Invalid timestamp, using current time instead.")
		return time.Now().UnixNano()
	}
	panic("Unreachable")
}
