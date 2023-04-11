package internal

import (
	"github.com/gin-gonic/gin"
)

func SetRouters(router *gin.Engine) {
	router.POST("/register", RegisterUser)
	router.POST("/login", LoginUser)
	router.POST("/loginBar", LoginBar)
	router.POST("/registerBar", RegisterBar)
	router.GET("/getUser", GetUser)
	router.GET("/getAllProducts", GetAllProducts)
	router.GET("/getAllFoods", GetAllFoods)
	router.PUT("/changeUser", ChangeUser)
}
