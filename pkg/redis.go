package pkg

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

var ClientRedis *redis.Client
var Ctx = context.Background()
var WebsocketChannel = "websocket_channel"

const (
	MaxReconnectAttempts = 5
	ReconnectInterval    = 5 * time.Second
)

func ReconnectToRedis() error {
	attempts := 0
	for attempts < MaxReconnectAttempts {
		attempts++

		if ClientRedis != nil {
			ClientRedis.Close()
		}

		time.Sleep(ReconnectInterval)

		ClientRedis = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		_, err := ClientRedis.Ping(Ctx).Result()
		if err == nil {
			return nil
		}
	}

	return errors.New("maximum reconnection attempts reached")
}

func init() {
	ClientRedis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
