# distributed-cron

分布式定时任务，使用分布式锁实现

## How to use

```go
    d := NewDistributedTask(redisClient)
    d.RegisterTasks(&HelloTask{})
    d.Start()
```

## Implement LockHasExpired interface 

```go
type LockHasExpired interface {
	Lock(key string, value interface{}, expiration time.Duration) error
	UnLock(key string, value interface{}) (interface{}, error)
    Expire(key string, value interface{}, expiration time.Duration) (interface{}, error)
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

func (r *Client) Lock(key string, value interface{}, expiration time.Duration) error {
	ok, err := r.redisClient.SetNX(context.Background(), key, value, expiration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("lock failed, key already use")
	}
	return nil
}

func (r *Client) UnLock(key string, value interface{}) (interface{}, error) {
	res, err := r.redisClient.Eval(context.Background(), UnlockScript, []string{key}, value).Result()
	if err != nil {
		logrus.Error("redis execute script fail", err.Error())
	}
	return res, err
}

// Expire RedisClient `expire` command
func (r *Client) Expire(key string, value interface{}, expiration time.Duration) (interface{}, error) {
    res, err := r.redisClient.Eval(context.Background(), expireScript, []string{key}, value, int(expiration/time.Millisecond)).Result()
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

Let's implement db Lock

```go
TODO
```
