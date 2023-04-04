package main

import (
	"TwistAndWrapS/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	internal.SetRouters(router)
	router.Run("localhost:8080")
}
