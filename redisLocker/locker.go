package redisLocker

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type Locker struct {
	key         string
	expire      time.Duration
	unlock      bool
	redisClient *redis.Client
	incrScript  *redis.Script
}

const incrLua = `
if redis.call('get', KEYS[1]) == ARGV[1] then
  return redis.call('expire', KEYS[1],ARGV[2]) 				
 else
   return '0' 					
end`

func NewLocker(redisClient *redis.Client, key string) *Locker {
	//默认30秒过期时间
	return &Locker{
		key:         key,
		expire:      time.Second * 30,
		incrScript:  redis.NewScript(incrLua),
		redisClient: redisClient,
	} //默认30秒
}

func (this *Locker) ping() {
	cmd := this.redisClient.Ping(context.Background())
	if cmd.Err() != nil {
		panic("redis connection error ")
	}

}

//有过期时间
func NewLockerWithTTL(redisClient *redis.Client, key string, expire time.Duration) *Locker {
	if expire.Seconds() <= 0 {
		panic("error expire")
	}
	return &Locker{
		key:         key,
		expire:      expire,
		incrScript:  redis.NewScript(incrLua),
		redisClient: redisClient,
	}
}
func (this *Locker) Lock() *Locker {
	this.ping()
	boolCmd := this.redisClient.SetNX(context.Background(), this.key, "1", this.expire)
	if ok, err := boolCmd.Result(); err != nil || !ok {
		panic(fmt.Sprintf("lock error with key:%s", this.key))
	}
	this.expandLockTime()
	fmt.Println("加锁成功")
	return this
}
func (this *Locker) expandLockTime() {
	sleepTime := this.expire.Seconds() * 2 / 3
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(sleepTime))
			if this.unlock {
				break
			}
			this.resetExpire()
		}
	}()
}

//重新设置过期时间
func (this *Locker) resetExpire() {
	cmd := this.incrScript.Run(context.Background(), this.redisClient, []string{this.key}, 1, this.expire.Seconds())
	v, err := cmd.Result()
	log.Printf("key=%s ,续期结果:%v,%v\n", this.key, err, v)
}

func (this *Locker) Unlock() {
	this.unlock = true
	this.redisClient.Del(context.Background(), this.key)
	fmt.Println("解锁成功")
}
