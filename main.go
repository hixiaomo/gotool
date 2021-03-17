package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hixiaomo/gotool/redisLocker"
	"time"
)

func main() {
	lock()
}

func lock() {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	locker := redisLocker.NewLockerWithTTL(redisClient, "ok", time.Second*3).Lock()
	defer locker.Unlock()
	time.Sleep(time.Second * 10)
	fmt.Print("ok")
}
