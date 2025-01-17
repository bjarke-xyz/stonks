package pkg

import (
	"time"
)

func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	location := t.Location()
	dayEndTime := time.Date(year, month, day, 23, 59, 59, 0, location)
	return dayEndTime
}
