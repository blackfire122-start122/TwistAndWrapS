package pkg

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var ClientRedis *redis.Client
var Ctx = context.Background()
var WebsocketChannel = "websocket_channel"

func init() {
	ClientRedis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
