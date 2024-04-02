package goschedule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CronSpecService struct {
}

func NewCronSpecService() SpecService {
	return CronSpecService{}
}

func (c CronSpecService) NextTime(curTime time.Time, express string) (*time.Time, *time.Duration, error) {
	if len(express) == 0 {
		return nil, nil, errors.New("表达式为空")
	}
	if strings.HasPrefix(express, "@every") {
		express2 := strings.TrimPrefix(express, "@every")
		express2 = strings.TrimSpace(express2)
		return c.every(curTime, express2)

	} else if strings.HasPrefix(express, "@day") {
		express2 := strings.TrimPrefix(express, "@day")
		express2 = strings.TrimSpace(express2)
		return c.day(curTime, express2)

	} else if strings.HasPrefix(express, "@year") {
		express2 := strings.TrimPrefix(express, "@year")
		express2 = strings.TrimSpace(express2)
		return c.year(curTime, express2)
	}
	return nil, nil, fmt.Errorf("表达式[%s]格式不正确", express)
}

func (c CronSpecService) every(curTime time.Time, timePeriod string) (*time.Time, *time.Duration, error) {
	if strings.HasSuffix(timePeriod, "s") {
		timeValueStr := strings.TrimSuffix(timePeriod, "s")
		timeValue, err := strconv.Atoi(timeValueStr)
		if err != nil {
			return nil, nil, fmt.Errorf("表达式[%s]格式不正确", timePeriod)
		}

		dur := time.Duration(int64(time.Second) * int64(timeValue))
		nextTime := curTime.Add(dur)
		return &nextTime, &dur, nil

	} else if strings.HasSuffix(timePeriod, "m") {
		timeValueStr := strings.TrimSuffix(timePeriod, "s")
		timeValue, err := strconv.Atoi(timeValueStr)
		if err != nil {
			return nil, nil, fmt.Errorf("表达式[%s]格式不正确", timePeriod)
		}

		dur := time.Duration(int64(time.Minute) * int64(timeValue))
		nextTime := curTime.Add(dur)
		return &nextTime, &dur, nil

	} else if strings.HasSuffix(timePeriod, "h") {
		timeValueStr := strings.TrimSuffix(timePeriod, "s")
		timeValue, err := strconv.Atoi(timeValueStr)
		if err != nil {
			return nil, nil, fmt.Errorf("表达式[%s]格式不正确", timePeriod)
		}

		dur := time.Duration(int64(time.Hour) * int64(timeValue))
		nextTime := curTime.Add(dur)
		return &nextTime, &dur, nil
	}

	return nil, nil, fmt.Errorf("表达式[%s]格式不正确,s-表示秒,m-表示分钟,h-表示小时", timePeriod)
}

func (c CronSpecService) day(curTime time.Time, timeStr string) (*time.Time, *time.Duration, error) {
	date1Str := fmt.Sprintf("%s %s", formatDate(time.Now()), timeStr)
	date1, err := parseTime(date1Str)
	if err != nil {
		return nil, nil, err
	}

	if curTime.Before(date1) {
		//TODO 还可以执行
		return &date1, nil, nil
	}

	date2Str := fmt.Sprintf("%s %s", formatDate(time.Now().AddDate(0, 0, 1)), timeStr)
	date2, err := parseTime(date2Str)
	if err != nil {
		return nil, nil, err
	}

	return &date2, nil, nil
}

func (c CronSpecService) year(curTime time.Time, timeStr string) (*time.Time, *time.Duration, error) {
	date1Str := fmt.Sprintf("%d-%s", time.Now().Year(), timeStr)
	date1, err := parseTime(date1Str)
	if err != nil {
		return nil, nil, err
	}

	if curTime.Before(date1) {
		//TODO 还可以执行
		return &date1, nil, nil
	}

	date2Str := fmt.Sprintf("%d-%s", time.Now().Year()+1, timeStr)
	date2, err := parseTime(date2Str)
	if err != nil {
		return nil, nil, err
	}

	return &date2, nil, nil
}
