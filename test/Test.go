package main

import (
	"context"
	"fmt"
	"github.com/zhengweiye/gopool"
	"github.com/zhengweiye/goschedule"
	"math/rand"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	pool := gopool.NewPool(100, 100, ctx)
	timer := goschedule.NewTimer(pool, ctx)
	timer.Start()

	timer.AddJob("test1", "测试1", true, 10*time.Second, "@every 10s", jobFuc, map[string]any{
		"name": "张三",
	})

	timer.AddJob("test2", "测试2", true, 15*time.Second, "@every 15s", jobFuc, map[string]any{
		"name": "李四",
	})

	time.Sleep(16 * time.Second)
	cancel()

	for {

	}
}

func jobFuc(param map[string]any) (err error, result string) {
	rand.Seed(time.Now().UnixNano())
	t := rand.Intn(10)
	time.Sleep(time.Duration(int64(time.Second) * int64(t)))

	fmt.Println("执行：", param["name"], time.Now())
	return
}
