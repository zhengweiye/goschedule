package goschedule

import "time"

type SpecService interface {
	NextTime(curTime time.Time, express string) (*time.Time, *time.Duration, error)
}

func newSpecService() SpecService {
	return NewCronSpecService()
}
