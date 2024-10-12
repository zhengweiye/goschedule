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

func (c CronSpecService) NextTime(isFirst bool, delay time.Duration, curTime time.Time, express string, missExec bool) (nextTime time.Time, err error) {
	if len(express) == 0 {
		err = errors.New("表达式为空")
		return
	}
	if strings.HasPrefix(express, "@every") {
		return c.every(delay, curTime, strings.TrimSpace(express[6:len(express)-1]), express[len(express)-1:])

	} else if strings.HasPrefix(express, "@day") {
		return c.day(isFirst, curTime, strings.TrimSpace(express[4:]), missExec)

	} else if strings.HasPrefix(express, "@week") {
		return c.week(isFirst, curTime, strings.TrimSpace(express[5:]), missExec)

	} else if strings.HasPrefix(express, "@month") {
		return c.month(isFirst, curTime, strings.TrimSpace(express[6:]), missExec)

	} else if strings.HasPrefix(express, "@year") {
		return c.year(isFirst, curTime, strings.TrimSpace(express[5:]), missExec)

	} else {
		err = fmt.Errorf("表达式[%s]格式不正确", express)
	}
	return
}

func (c CronSpecService) every(delay time.Duration, curTime time.Time, timePeriod, timeUnit string) (nextTime time.Time, err error) {
	//fmt.Println(">>>>>>>>>>>>>>>>>>>timePeriod=", timePeriod, ", timeUnit=", timeUnit)
	if delay > 0 {
		curTime = curTime.Add(delay)
	}
	timeValue, err2 := strconv.Atoi(timePeriod)
	if err2 != nil {
		err = fmt.Errorf("时间间隔[%s]格式不正确", timePeriod)
		return
	}

	switch timeUnit {
	case "s":
		nextTime = curTime.Add(time.Duration(int64(time.Second) * int64(timeValue)))

	case "m":
		nextTime = curTime.Add(time.Duration(int64(time.Minute) * int64(timeValue)))

	case "h":
		nextTime = curTime.Add(time.Duration(int64(time.Hour) * int64(timeValue)))

	default:
		err = fmt.Errorf("时间间隔单位[%s]格式不正确,s-表示秒,m-表示分钟,h-表示小时", timePeriod)
	}
	return
}

func (c CronSpecService) day(isFirst bool, curTime time.Time, timeStr string, missExec bool) (nextTime time.Time, err error) {
	//fmt.Println(">>>>>>>>>>>timeStr=", timeStr)
	if isFirst {
		nextTime, err = getNewDayTime(curTime, 0, timeStr)
		if err != nil {
			return
		}
		if nextTime.Before(curTime) {
			if missExec {
				return
			}
		} else {
			return
		}
	}

	nextTime, err = getNewDayTime(curTime, 1, timeStr)
	return
}

// timeStr格式：monday 09:00:00
func (c CronSpecService) week(isFirst bool, curTime time.Time, timeStr string, missExec bool) (nextTime time.Time, err error) {
	if isFirst {
		// 获取本周的执行时间
		nextTime, err = getNewWeekTime(curTime, 0, timeStr)
		if err != nil {
			return
		}
		if nextTime.Before(curTime) { // 本周执行时间已经过了
			if missExec { // 错过执行，补偿执行
				return
			}
			// 错过执行，那么获取下周的时间
		} else { // 本周执行时间没有过
			return
		}
	}

	// 获取下周的执行时间
	nextTime, err = getNewDayTime(curTime, 1, timeStr)
	return
}

func (c CronSpecService) month(isFirst bool, curTime time.Time, timeStr string, missExec bool) (nextTime time.Time, err error) {
	return
}

func (c CronSpecService) year(isFirst bool, curTime time.Time, timeStr string, missExec bool) (nextTime time.Time, err error) {
	if isFirst {
		nextTime, err = getNewYearTime(curTime, 0, timeStr)
		if err != nil {
			return
		}
		if nextTime.Before(curTime) {
			if missExec {
				return
			}
		} else {
			return
		}
	}

	nextTime, err = getNewYearTime(curTime, 1, timeStr)
	return
}
