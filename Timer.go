package goschedule

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"
)

type JobFunc func(param map[string]any) (error, string)
type LogFunc func(log Log)

var signalChan = make(chan os.Signal, 1)

type Timer struct {
	jobs        []*Job
	scheduler   *time.Timer
	specService SpecService
	jobLock     sync.RWMutex
	defaultTime time.Duration
	wg          *sync.WaitGroup
	httpServer  *http.Server
	signalFile  string
	logFunc     LogFunc
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

var schedule *Timer
var scheduleNewOnce sync.Once
var scheduleStartOnce sync.Once

func NewTimer() *Timer {
	scheduleNewOnce.Do(func() {
		schedule = &Timer{
			defaultTime: 10 * time.Second,
			wg:          &sync.WaitGroup{},
			specService: newSpecService(),
		}
	})
	return schedule
}

func init() {
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGILL,
		syscall.SIGTRAP,
		syscall.SIGABRT,
		syscall.SIGBUS,
		syscall.SIGFPE,
		syscall.SIGKILL,
		syscall.SIGSEGV,
		syscall.SIGPIPE,
		syscall.SIGALRM,
		syscall.SIGTERM, //terminated ==> docker重启项目会触发该信号
	)
}

/**
 * 设置日志构造函数
 */

func (t *Timer) SetLogFunc(fun LogFunc) {
	t.logFunc = fun
}

/**
 * 监听到信号量时，打印日志路径
 */

func (t *Timer) SetSignalFile(file string) {
	t.signalFile = file
}

/**
 * 设置http服务，用于优雅听证
 */

func (t *Timer) SetHttpServer(server *http.Server) {
	t.httpServer = server
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
	scheduleStartOnce.Do(func() {
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
		case signalCmd := <-signalChan:
			switch signalCmd {
			case os.Interrupt:
				t.stop(signalCmd)
			case os.Kill:
				t.stop(signalCmd)
			case syscall.SIGTERM:
				t.stop(signalCmd)
			}
		}
	}
}

func (t *Timer) stop(signalCmd os.Signal) {
	// 停止定时器
	t.scheduler.Stop()
	// 停止监听信号量
	signal.Stop(signalChan)

	// 记录到文档里面
	if len(t.signalFile) > 0 {
		fileName := t.signalFile
		os.MkdirAll(filepath.Dir(fileName), 0750)
		file, _ := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		defer file.Close()
		writer := bufio.NewWriter(file)
		writer.WriteString(fmt.Sprintf("时间=%s,指令=%s\r\n", formatTime(time.Now()), signalCmd.String()))
		writer.Flush()
	}

	// 关闭http
	if t.httpServer != nil {
		// 等所有任务执行完成，才停止服务
		t.wg.Wait()

		// 关闭http服务
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		t.httpServer.Shutdown(ctx)
	}
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
	//TODO #issue, 集成线程池
	logJobs := make([]*LogJob, len(execJobs))
	for execIndex, execJob := range execJobs {
		t.wg.Add(1)
		go func(index int, job *Job) {
			err, resultMsg := t.execJob(job)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			logJobs[index] = &LogJob{
				key:        job.key,
				name:       job.name,
				ExecTime:   timeNowStr,
				ExecErr:    errMsg,
				ExecResult: resultMsg,
				NextTime:   formatTime(job.nextTime),
			}
		}(execIndex, execJob)
	}

	// 等job执行完成, 并且更新job的下次执行时间
	//TODO #issue, 感觉没必要等所有执行的job完成才进行下一轮，这样会导致其它定时器的时间准确性
	t.wg.Wait()

	// 保存日志
	if t.logFunc != nil {
		var cronLog Log
		cronLog.ExecTime = timeNow
		cronLog.Jobs = logJobs

		go func(cronLog2 Log) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(">>> insert into cron_log err:", err)
				}
			}()
			t.logFunc(cronLog2)
		}(cronLog)
	}

	// 重置定时器
	duration := t.getLatestDuration()
	fmt.Println(">>>重置定时器时间:", duration.Seconds(), "秒")
	t.scheduler.Reset(duration)
}

func (t *Timer) getLatestDuration() time.Duration {
	d := t.defaultTime
	if len(t.jobs) > 0 {
		timeNow := time.Now()
		firstJobNextTime := t.jobs[0].nextTime
		if firstJobNextTime.Before(timeNow) || firstJobNextTime.Equal(timeNow) {
			d = 500 * time.Millisecond // 已经错过的，则立马执行
		} else if firstJobNextTime.After(time.Now()) {
			d = t.jobs[0].nextTime.Sub(time.Now())
		}
	}
	return d
}

func (t *Timer) execJob(job *Job) (err error, result string) {
	defer func() {
		defer t.wg.Done()
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("%v", err2)
			fmt.Printf(">>> %s执行异常:%v\n", job.name, err)
		}
		// 计算下次时间
		if job.period != nil {
			job.nextTime = time.Now().Add(*job.period)
		} else {
			nextTime, _, err2 := t.specService.NextTime(time.Now(), job.cronExpress)
			if err2 != nil {
				fmt.Println(job.name, ">>获取NextTime错误=", err2)
				return
			}
			job.nextTime = *nextTime
		}
	}()

	err, result = job.jobFunc(job.param)
	return
}
