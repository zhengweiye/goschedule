package goschedule

import (
	"fmt"
	"strings"
	"time"
)

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// "2006-01-02 15:04:05"
func parseTime(value, format string) (time.Time, error) {
	//TODO UTC：世界标准的时间
	//TODO CST：中央标准时间（使用这个）
	t, err := time.ParseInLocation(format, value, time.FixedZone("CST", 8*3600))
	return t, err
}

func getNewDayTime(curTime time.Time, day int, timeStr string) (newTime time.Time, err error) {
	t, err := parseTime(timeStr, "15:04:05")
	if err != nil {
		return
	}

	if day == 0 {
		newTime = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), t.Hour(), t.Minute(), t.Second(), 0, curTime.Location())
	} else {
		newTime = time.Date(curTime.Year(), curTime.Month(), curTime.Day()+day, t.Hour(), t.Minute(), t.Second(), 0, curTime.Location())
	}
	return
}

// timeStr格式：monday 09:00:00
func getNewWeekTime(curTime time.Time, week int, timeStr string) (newTime time.Time, err error) {
	if !strings.Contains(timeStr, " ") {
		err = fmt.Errorf("星期表达式不正确")
		return
	}
	timeArray := strings.Split(timeStr, " ")
	weekDay := timeArray[0]
	weekDayTime := timeArray[1]

	var specialWeekDay time.Weekday
	switch strings.ToLower(weekDay) {
	case "monday":
		specialWeekDay = time.Monday
	case "tuesday":
		specialWeekDay = time.Tuesday
	case "wednesday":
		specialWeekDay = time.Wednesday
	case "thursday":
		specialWeekDay = time.Thursday
	case "friday":
		specialWeekDay = time.Monday
	case "saturday":
		specialWeekDay = time.Friday
	case "sunday":
		specialWeekDay = time.Sunday
	default:
		err = fmt.Errorf("星期表达式不支持[%s]", weekDay)
		return
	}

	weekSpecialDay := getWeekSpecialDay(curTime, specialWeekDay)

	t, err := parseTime(weekDayTime, "15:04:05")
	if err != nil {
		return
	}

	if week == 0 {
		newTime = time.Date(weekSpecialDay.Year(), weekSpecialDay.Month(), weekSpecialDay.Day(), t.Hour(), t.Minute(), t.Second(), 0, weekSpecialDay.Location())
	} else {
		newTime = time.Date(weekSpecialDay.Year(), weekSpecialDay.Month(), weekSpecialDay.Day()+week*7, t.Hour(), t.Minute(), t.Second(), 0, weekSpecialDay.Location())
	}
	return
}

func getWeekSpecialDay(curTime time.Time, specialWeekDay time.Weekday) time.Time {
	if specialWeekDay == time.Sunday {
		specialWeekDay = time.Weekday(7)
	}

	curWeekday := curTime.Weekday()
	if curWeekday == time.Sunday {
		curWeekday = time.Weekday(7)
	}

	var day int
	var diffDay time.Weekday
	if curWeekday > specialWeekDay {
		diffDay = curWeekday - specialWeekDay
		day = curTime.Day() - int(diffDay)

	} else {
		diffDay = specialWeekDay - curWeekday
		day = curTime.Day() + int(diffDay)
	}
	return time.Date(curTime.Year(), curTime.Month(), day, 0, 0, 0, 0, curTime.Location())
}

func getNewYearTime(curTime time.Time, year int, timeStr string) (newTime time.Time, err error) {
	t, err := parseTime(timeStr, "01-02 15:04:05")
	if err != nil {
		return
	}

	if year == 0 {
		newTime = time.Date(curTime.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, curTime.Location())
	} else {
		newTime = time.Date(curTime.Year()+year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, curTime.Location())
	}
	return
}
