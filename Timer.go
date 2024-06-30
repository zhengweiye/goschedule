package goschedule

import (
	"context"
	"fmt"
	"github.com/zhengweiye/gopool"
	"sort"
	"sync"
	"time"
)

type JobFunc func(param map[string]any) (err error, result string)
type LogFunc func(log Log)

type Timer struct {
	jobs        []*Job
	scheduler   *time.Timer
	specService SpecService
	jobLock     sync.RWMutex
	defaultTime time.Duration
	logFunc     LogFunc
	pool        *gopool.Pool
	wg          *sync.WaitGroup
	ctx         context.Context
	quit        chan bool
	isShutdown  bool
}

type Job struct {
	key         string
	name        string
	jobFunc     JobFunc
	cronExpress string         // cron表达式
	period      *time.Duration // 时间间隔
	nextTime    time.Time
	param       map[string]any
}

type Log struct {
	ExecTime time.Time
	Jobs     []*LogJob
}

type LogJob struct {
	key        string
	name       string
	ExecTime   string
	ExecErr    string
	ExecResult string
	NextTime   string
}

var timerObj *Timer
var timerOnce sync.Once
var timerStartOnce sync.Once

func NewTimer(pool *gopool.Pool, ctx context.Context) *Timer {
	timerOnce.Do(func() {
		timerObj = &Timer{
			defaultTime: 10 * time.Second,
			wg:          &sync.WaitGroup{},
			specService: newSpecService(),
			ctx:         ctx,
			pool:        pool,
			quit:        make(chan bool),
			isShutdown:  false,
		}
		fmt.Printf(">>>>>>[定时器] 线程池指针：%p\n", pool)
	})
	return timerObj
}

/**
 * 设置日志构造函数
 */

func (t *Timer) SetLogFunc(fun LogFunc) {
	t.logFunc = fun
}

/**
 * 给定时器添加定时作业
 * jobKey: 作业标识
 * jobName: 作业名称
 * existReplace: jobKey存在时,是否替换旧的job
 * delay: 延迟执行时间
 * cronExpress: 执行周期
 * jobFunc: 执行的业务函数
 * param: 执行时的参数
 */

func (t *Timer) AddJob(jobKey, jobName string, existReplace bool,
	delay time.Duration, cronExpress string,
	jobFunc JobFunc, param map[string]any) {

	t.jobLock.Lock()
	defer t.jobLock.Unlock()

	if !existReplace {
		for _, job := range t.jobs {
			if job.key == jobKey {
				panic(fmt.Errorf("jobKey=%s已经存在", jobKey))
			}
		}
	} else {
		for _, job := range t.jobs {
			if job.key == jobKey {
				t.delJob(job)
			}
		}
	}

	// 情况一：先AddJob-->Start ===> 没有问题
	// 情况二：先Start-->AddJob ===> 下一个周期会被执行
	// 情况三：先AddJob-->Start-->AddJob ===> 下一个周期会被执行
	_, period, err := t.specService.NextTime(time.Now(), cronExpress)
	nextTime := time.Now().Add(delay)
	if err != nil {
		panic(fmt.Errorf("%s的表达式错误", jobName))
	}

	//fmt.Println(">>>AddJob() ", ", joName=", jobName, ", cron=", cronExpress, ", 预计执行时间=", nextTime.Format("2006-01-02 15:04:05"))
	t.addJob(&Job{
		key:         jobKey,
		name:        jobName,
		jobFunc:     jobFunc,
		cronExpress: cronExpress,
		period:      period,
		nextTime:    nextTime,
		param:       param,
	})
}

func (t *Timer) RemoveJob(jobKey string) {
	t.jobLock.Lock()
	defer t.jobLock.Unlock()

	var jobObj *Job
	for _, job := range t.jobs {
		if job.key == jobKey {
			jobObj = job
			break
		}
	}
	if jobObj != nil {
		t.delJob(jobObj)
	}
}

func (t *Timer) addJob(job *Job) {
	t.jobs = append(t.jobs, job)
	sort.Slice(t.jobs, func(i, j int) bool {
		return t.jobs[i].nextTime.Before(t.jobs[j].nextTime)
	})
}

func (t *Timer) delJob(job *Job) {
	for index, item := range t.jobs {
		if item.key == job.key {
			t.jobs = append(t.jobs[:index], t.jobs[index+1:]...)
		}
	}

	sort.Slice(t.jobs, func(i, j int) bool {
		return t.jobs[i].nextTime.Before(t.jobs[j].nextTime)
	})
}

/**
 * 启动定时器
 */

func (t *Timer) Start() {
	timerStartOnce.Do(func() {
		// 创建定时器
		t.scheduler = time.NewTimer(t.getLatestDuration())

		// 监听
		go t.listen()
	})
}

func (t *Timer) listen() {
	for {
		select {
		case <-t.scheduler.C:
			t.process()

		case <-t.quit:
			fmt.Println("[定时器] 退出定时器监听......................")
			return

		case <-t.ctx.Done():
			fmt.Println("[定时器] 接受到Context取消信号...............")
			if !t.isShutdown {
				t.stop()
			}
		}
	}
}

func (t *Timer) stop() {
	// 停止定时器
	t.isShutdown = true
	t.scheduler.Stop()

	fmt.Println("[定时器] 关闭定时器, 等待结束..........")

	t.wg.Wait()
	close(t.quit)

	fmt.Println("[定时器] 关闭定时器, 已经结束..........")
}

func (t *Timer) process() {
	timeNow := time.Now()
	timeNowStr := formatTime(timeNow)

	// 获取需要执行的job
	t.jobLock.RLock()
	defer t.jobLock.RUnlock()

	execJobs := []*Job{}
	for _, job := range t.jobs {
		if job.nextTime.Before(timeNow) || job.nextTime.Equal(timeNow) {
			execJobs = append(execJobs, job)
		}
	}

	// 执行job
	logJobs := make([]*LogJob, len(execJobs))
	for execIndex, execJob := range execJobs {
		t.wg.Add(1)

		// 协程池执行任务
		futureChan := make(chan gopool.Future)
		//close(futureChan)// 不需要关闭，协程池里面已经关闭了

		t.pool.ExecTaskFuture(gopool.JobFuture{
			JobName: "执行定时任务",
			JobFunc: t.execJob,
			JobParam: map[string]any{
				"job": execJob,
			},
			Future: futureChan,
		})

		// 获取协程池执行结果
		future := <-futureChan

		// 执行结果处理
		var errMsg string
		var resultMsg string
		var nextTime time.Time

		if future.Error != nil {
			errMsg = future.Error.Error()
		}
		if future.Result != nil {
			execResult, ok := future.Result.(ExecResult)
			if ok {
				resultMsg = execResult.Result
				nextTime = execResult.NextTime
			}
		}
		logJobs[execIndex] = &LogJob{
			key:        execJob.key,
			name:       execJob.name,
			ExecTime:   timeNowStr,
			ExecErr:    errMsg,
			ExecResult: resultMsg,
			NextTime:   formatTime(nextTime),
		}
	}

	// 等job执行完成, 并且更新job的下次执行时间
	//TODO #issue, 感觉没必要等所有执行的job完成才进行下一轮，这样会导致其它定时器的时间准确性
	t.wg.Wait()

	// 保存日志
	if t.logFunc != nil {
		var cronLog Log
		cronLog.ExecTime = timeNow
		cronLog.Jobs = logJobs

		t.pool.ExecTask(gopool.Job{
			JobName: "定时器执行日志保存",
			JobFunc: t.logJob,
			JobParam: map[string]any{
				"log": cronLog,
			},
		})
	}

	// 重置定时器
	if !t.isShutdown {
		duration := t.getLatestDuration()
		fmt.Println(">>> [定时器] 重置定时器时间：", duration.Seconds(), "秒")
		t.scheduler.Reset(duration)
	}
}

func (t *Timer) getLatestDuration() time.Duration {
	if len(t.jobs) == 0 {
		return t.defaultTime
	}

	sort.Slice(t.jobs, func(i, j int) bool {
		return t.jobs[i].nextTime.Before(t.jobs[j].nextTime)
	})

	timeNow := time.Now()
	firstJobNextTime := t.jobs[0].nextTime
	if firstJobNextTime.Before(timeNow) || firstJobNextTime.Equal(timeNow) {
		return 500 * time.Millisecond // 已经错过的，则立马执行
	} else {
		return firstJobNextTime.Sub(timeNow)
	}
}

func (t *Timer) logJob(workerId int, jobName string, param map[string]any) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">>> [定时器] 执行日志报错异常：", err)
		}
	}()
	t.logFunc(param["log"].(Log))
	return nil
}

type ExecResult struct {
	Result   string
	NextTime time.Time
}

func (t *Timer) execJob(workerId int, jobName string, param map[string]any, future chan gopool.Future) {
	//TODO 感觉放这里不合适，因为process()循环里面，下面还有代码执行
	defer t.wg.Done()

	// 参数
	job := param["job"].(*Job)
	var execError error
	var execResult string

	// 更新job下一轮时间
	// 异常处理
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf(">>> [定时器] [%s] 执行任务异常：%v\n", job.name, err)
		}

		if job.period != nil {
			job.nextTime = time.Now().Add(*job.period)
		} else {
			nextTime, _, err := t.specService.NextTime(time.Now(), job.cronExpress)
			if err != nil {
				fmt.Printf(">>> [定时器] [%s] 更新job.NextTime异常：%v\n", job.name, err)
				return
			}
			job.nextTime = *nextTime
		}

		future <- gopool.Future{
			Error: execError,
			Result: ExecResult{
				Result:   execResult,
				NextTime: job.nextTime,
			},
		}
	}()

	// 执行任务
	execError, execResult = job.jobFunc(job.param)
	if execError != nil {
		panic(execError)
	}
}
