package dcron

import (
	"context"
	"errors"
	v8 "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	RClient      *Client
	unlockScript = UnlockScript
	expireScript = ExpireScript
)

func Init() {
	redisClient := v8.NewClient(&v8.Options{
		Addr:     "127.0.0.1:6379",
		Username: "",
		Password: "12345",
	})
	RClient = New(redisClient)
}

type Client struct {
	redisClient v8.UniversalClient
}

func New(redisClient v8.UniversalClient) *Client {
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
	res, err := r.redisClient.Eval(context.Background(), unlockScript, []string{key}, value).Result()
	if err != nil {
		logrus.Error("redis execute unlock script fail", err.Error())
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
