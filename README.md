# distributed-cron

分布式定时任务，使用分布式锁实现

## How to use

```go
    d := NewDistributedTask(redisClient, "YOUR_NAME_SPACE", "YOUR_SERVER_NAME")
    d.RegisterTasks(&HelloTask{})
    d.Start()
```

## Task Mode

```go
ModeDefault             Mode = iota //每到执行时间直接执行，不管前面任务执行是否完成
ModeSkipIfStillRunning              //到了执行时间，前面任务执行未完成，则直接跳过，本次不执行，等待下次执行时间
ModeDelayIfStillRunning             //到了执行时间，前面任务执行未完成，则等待其完成后立即执行(可能造成任务堆积，不建议使用)
```

## Implement LockHasExpired interface 

```go
type LockHasExpired interface {
	Lock(key string, value interface{}, expiration time.Duration) (bool, error)
	UnLock(key string, value interface{}) (interface{}, error)
    ExpireWithVal(key string, value interface{}, expiration time.Duration) (interface{}, error)
	TTL(key string) (time.Duration, error)
}
```

Let's implement redis Lock

```go
import (
	"context"
	"errors"
	redis "github.com/go-redis/redis/redis"
	"github.com/sirupsen/logrus"
	"time"
)

var RClient *Client

func init() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "",
		Password: "",
	})
	RClient = New(redisClient)
}

type Client struct {
	redisClient redis.UniversalClient
}

func New(redisClient redis.UniversalClient) *Client {
	return &Client{
		redisClient: redisClient,
	}
}

func (r *Client) Close() {
	_ = r.redisClient.Close()
}

func (r *Client) Lock(key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.redisClient.SetNX(context.Background(), key, value, expiration).Result()
}

func (r *Client) UnLock(key string, value interface{}) (interface{}, error) {
	res, err := r.redisClient.Eval(context.Background(), UnlockScript, []string{key}, value).Result()
	if err != nil {
		logrus.Error("redis execute script fail", err.Error())
	}
	return res, err
}

// Expire RedisClient `expire` command
func (r *Client) ExpireWithVal(key string, value interface{}, expiration time.Duration) (interface{}, error) {
    res, err := r.redisClient.Eval(context.Background(), ExpireScript, []string{key}, value, int(expiration/time.Millisecond)).Result()
    if err != nil {
        logrus.Error("redis execute expire script fail, ", err.Error())
    }
    return res, err
}

func (r *Client) TTL(key string) (time.Duration, error) {
	keys := r.redisClient.TTL(context.Background(), key)
	return keys.Result()
}
```
