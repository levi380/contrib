package helper

import "time"

func TodaySt(loc *time.Location) time.Time {
	year, month, day := time.Now().In(loc).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func MonthSt(loc *time.Location) time.Time {
	year, month, _ := time.Now().In(loc).Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, loc)
}
