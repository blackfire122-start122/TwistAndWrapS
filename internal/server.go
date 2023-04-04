package internal

import (
	"github.com/gin-gonic/gin"
)

func SetRouters(router *gin.Engine) {
	router.POST("/register", RegisterUser)
	router.POST("/login", LoginUser)
	router.GET("/getUser", GetUser)
	router.GET("/getAllProducts", GetAllProducts)
}
