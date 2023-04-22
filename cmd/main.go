package main

import (
	"TwistAndWrapS/internal"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	internal.SetRouters(router)
	go internal.Broadcaster()
	err := router.Run("localhost:8080")
	if err != nil {
		fmt.Println(err)
	}
}
