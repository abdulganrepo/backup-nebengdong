package date

import (
	"time"
)

func CurrentUTCTime() *time.Time {
	// Local time
	/* currentTime := time.Now()
	time := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond(), time.UTC)

	return &time */

	// UTC
	currentTime := time.Now().UTC()
	return &currentTime
}
