package goschedule

import "time"

type SpecService interface {
	NextTime(isFirst bool, delay time.Duration, curTime time.Time, express string, missExec bool) (nextTime time.Time, err error)
}

func newSpecService() SpecService {
	return NewCronSpecService()
}
