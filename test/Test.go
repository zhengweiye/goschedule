package main

import (
	"fmt"
	"github.com/zhengweiye/goschedule"
	"time"
)

func main() {
	timer := goschedule.NewTimer()

	timer.AddJob("test", "测试", true, time.Second, "@every 5s", jobFuc, map[string]any{
		"name": "张三",
	})
	timer.AddJob("test", "测试", true, time.Second, "@every 5s", jobFuc, map[string]any{
		"name": "张三",
	})
	timer.SetLogFunc(logRecord)
	timer.Start()

	time.Sleep(10 * time.Second)
}

func logRecord(log goschedule.Log) {
	fmt.Println("日志：", log)
}

func jobFuc(param map[string]any) (err error, result string) {
	fmt.Println("执行：", param)
	result = "执行成功"
	return
}
