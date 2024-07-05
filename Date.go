package goschedule

import "time"

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
