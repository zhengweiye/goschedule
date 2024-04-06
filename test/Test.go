package main

import (
	"fmt"
	"github.com/zhengweiye/goschedule"
	"time"
)

func main() {
	timer := goschedule.NewTimer()
	timer.Start()

	timer.AddJob("test1", "测试1", true, 10*time.Second, "@every 30s", jobFuc, map[string]any{
		"name": "张三",
	})

	timer.AddJob("test2", "测试2", true, 20*time.Second, "@every 40s", jobFuc, map[string]any{
		"name": "李四",
	})

	for {

	}
}

func jobFuc(param map[string]any) (err error, result string) {
	fmt.Println("执行：", param["name"], time.Now())
	return
}
