package internal

import (
	"github.com/gin-gonic/gin"
)

func SetRouters(router *gin.Engine) {
	router.POST("api/user/register", RegisterUser)
	router.POST("api/user/login", LoginUser)
	router.GET("api/user/logout", LogoutUser)
	router.GET("api/user/getUser", GetUser)
	router.PUT("api/user/changeUser", ChangeUser)
	router.POST("api/user/orderFood", OrderFood)
	router.GET("api/user/getAllWorkedBars", GetAllWorkedBars)
	router.GET("api/user/getAllBars", GetAllBars)
	router.GET("api/user/getTypes", GetTypes)
	router.GET("api/user/getAllFoods", GetAllFoods)

	router.POST("api/bar/loginBar", LoginBar)
	router.GET("api/bar/getAllProducts", GetAllProducts)
	router.GET("websocket/wsChat", WsChat)

	router.POST("api/admin/createProduct", CreateProduct)
	router.POST("api/admin/registerBar", RegisterBar)
	router.POST("api/admin/changeFood", ChangeFood)
	router.DELETE("api/admin/deleteFood/:id", DeleteFood)

}
