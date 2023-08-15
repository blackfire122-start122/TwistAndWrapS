package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

var ClientRedis *redis.Client
var Ctx = context.Background()
var WebsocketChannel = "websocket_channel"

var ReconnectInterval = 1 * time.Second

func ReconnectToRedis() error {
	for ReconnectInterval < 10*time.Minute {
		fmt.Println("try reconnect", ReconnectInterval)
		if ClientRedis != nil {
			ClientRedis.Close()
		}

		time.Sleep(ReconnectInterval)

		ClientRedis = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		res, err := ClientRedis.Ping(Ctx).Result()

		if err == nil {
			fmt.Println(ClientRedis, res)
			return nil
		}

		ReconnectInterval += ReconnectInterval
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
