package main

import "time"

var initNow = today()

const oneDay = 24 * time.Hour

func today() time.Time {
	now := time.Now().In(time.Local)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}
