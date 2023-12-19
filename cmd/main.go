package main

import (
	"TwistAndWrapS/internal"
	. "TwistAndWrapS/pkg/logging"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"strconv"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	var Port = int64(8080)
	for {
		if isPortAvailable(Port) {
			break
		}
		Port++
	}

	router := gin.Default()
	gin.SetMode(gin.DebugMode)
	router.Use(CORSMiddleware())
	internal.SetRouters(router)
	go internal.Broadcaster()
	go internal.RedisReceiver()

	err := router.Run("localhost:" + strconv.FormatInt(Port, 10))
	if err != nil {
		ErrorLogger.Println(err.Error())
	}
}

func isPortAvailable(port int64) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}
