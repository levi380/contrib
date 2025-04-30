package goment

import (
	"errors"
	"net/http"
	"sort"
	"time"
)

func Week(now time.Time, loc *time.Location, val int) (time.Time, time.Time) {

	offset := int(time.Monday - now.Weekday())
	//周日做特殊判断 因为time.Monday = 0
	if offset > 0 {
		offset = -6
	}

	y, m, d := now.In(loc).Date()
	thisWeek := time.Date(y, m, d, 0, 0, 0, 0, loc)

	st := thisWeek.AddDate(0, 0, offset+7*val)
	et := thisWeek.AddDate(0, 0, (offset + 6 + 7*val))

	return st, et.Add(time.Second * 86399)
}

func Month(now time.Time, loc *time.Location, mon int) (time.Time, time.Time) {

	y, m, _ := now.In(loc).Date()
	thisMonth := time.Date(y, m, 1, 0, 0, 0, 0, loc)

	st := thisMonth.AddDate(0, mon, 0)
	et := thisMonth.AddDate(0, mon+1, -1)
	return st, et.Add(time.Second * 86399)
}

func Day(now time.Time, loc *time.Location, val int) (time.Time, time.Time) {

	y, m, d := now.In(loc).Date()
	thisDay := time.Date(y, m, d, 0, 0, 0, 0, loc)

	st := thisDay.AddDate(0, 0, val)
	return st, st.Add(time.Second * 86399)
}

func SpecificDate(now time.Time, loc *time.Location, mon, s, e int) (time.Time, time.Time) {

	y, m, _ := now.In(loc).AddDate(0, mon, 0).Date()

	st := time.Date(y, m, s, 0, 0, 0, 0, loc)
	et := time.Date(y, m, e, 23, 59, 59, 0, loc)

	return st, et
}

/*
timestamps := []int64{
		1672531200, // 2023-01-01
		1672617600, // 2023-01-02
		1672704000, // 2023-01-03
	}

	if ConsecutiveDays(timestamps) {
		fmt.Println("这些时间戳是连续的天数")
	} else {
		fmt.Println("这些时间戳不是连续的天数")
	}
*/
//判断一组时间戳，是不是连续日期
func ConsecutiveDays(timestamps []int64) bool {
	if len(timestamps) < 2 {
		return true // 一个或零个时间戳，默认为连续
	}

	// 将时间戳按升序排序
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	// 遍历检查相邻的时间戳是否相差一天
	for i := 1; i < len(timestamps); i++ {
		t1 := time.Unix(timestamps[i-1], 0)
		t2 := time.Unix(timestamps[i], 0)

		// 计算两个时间戳的日期差
		diffDays := t2.Sub(t1).Hours() / 24
		if diffDays != 1 {
			return false // 如果日期差不是 1 天，则不是连续的
		}
	}
	return true
}

func StrToTime(value string, loc *time.Location) (time.Time, error) {

	var (
		t   time.Time
		err error
	)

	if value == "" {
		return t, errors.New("formatted time cannot be empty")
	}
	layouts := []string{
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05 -0700 MST",
		"2006/01/02 15:04:05 -0700",
		"2006/01/02 15:04:05",
		"2006-01-02 -0700 MST",
		"2006-01-02 -0700",
		"2006-01-02",
		"2006/01/02 -0700 MST",
		"2006/01/02 -0700",
		"2006/01/02",
		"2006-01-02 15:04:05 -0700 -0700",
		"2006/01/02 15:04:05 -0700 -0700",
		"2006-01-02 -0700 -0700",
		"2006/01/02 -0700 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		http.TimeFormat,
	}

	for _, layout := range layouts {
		t, err = time.ParseInLocation(layout, value, loc)
		if err == nil {
			return t, nil
		}
	}

	return t, errors.New("Invalid")
}
