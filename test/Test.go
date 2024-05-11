package main

import (
	"fmt"
	"github.com/zhengweiye/gopool"
	"github.com/zhengweiye/goschedule"
	"time"
)

func main() {
	timer := goschedule.NewTimer()
	timer.Start()
	timer.SetPool(gopool.NewPool(100, 100))

	timer.AddJob("test1", "测试1", true, 10*time.Second, "@every 30s", jobFuc, map[string]any{
		"name": "张三",
	})

	for {

	}
}

func jobFuc(param map[string]any) (err error, result string) {
	fmt.Println("执行：", param["name"], time.Now())
	return
}
