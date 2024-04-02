package goschedule

import "time"

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func parseTime(str string) (time.Time, error) {
	//TODO UTC：世界标准的时间
	//TODO CST：中央标准时间（使用这个）
	t, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.FixedZone("CST", 8*3600))
	return t, err
}
