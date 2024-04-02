package main

import (
	"fmt"
	"github.com/zhengweiye/goschedule"
	"time"
)

func main() {
	timer := goschedule.NewTimer()

	fmt.Println("AddJob:", time.Now())
	timer.AddJob("test", "测试", true, 10*time.Second, "@every 30s", jobFuc, map[string]any{
		"name": "张三",
	})
	timer.SetLogFunc(logRecord)
	timer.Start()

	time.Sleep(100 * time.Second)
}

func logRecord(log goschedule.Log) {
	fmt.Println("日志：", log)
}

func jobFuc(param map[string]any) (err error, result string) {
	fmt.Println("执行：", time.Now())
	result = "执行成功"
	return
}
