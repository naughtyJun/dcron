package example

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gitlab.bianjie.ai/spark/common-modules/dcron"
)

var RClient *Client

func init() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
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
	return r.redisClient.SetNX(key, value, expiration).Result()
}

func (r *Client) UnLock(key string, value interface{}) (interface{}, error) {
	res, err := r.redisClient.Eval(dcron.UnlockScript, []string{key}, value).Result()
	if err != nil {
		logrus.Error("redis execute script fail", err.Error())
	}
	return res, err
}

// Expire RedisClient `expire` command
func (r *Client) ExpireWithVal(key string, value interface{}, expiration time.Duration) (interface{}, error) {
	res, err := r.redisClient.Eval(dcron.ExpireScript, []string{key}, value, int(expiration/time.Millisecond)).Result()
	if err != nil {
		logrus.Error("redis execute expire script fail, ", err.Error())
	}
	return res, err
}

func (r *Client) TTL(key string) (time.Duration, error) {
	keys := r.redisClient.TTL(key)
	return keys.Result()
}
