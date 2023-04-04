package internal

import (
	. "TwistAndWrapS/pkg"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

func GetUser(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp := make(map[string]string)

	resp["Id"] = strconv.FormatUint(user.Id, 10)
	resp["Username"] = user.Username
	resp["Email"] = user.Email

	c.JSON(http.StatusOK, resp)
}

func RegisterUser(c *gin.Context) {
	resp := make(map[string]string)

	var user UserRegister
	bodyBytes, _ := io.ReadAll(c.Request.Body)

	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("here")

	if user.Password1 == "" || user.Username == "" {
		resp["Register"] = "Not all field"

		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if user.Password1 != user.Password2 {
		resp["Register"] = "Not equal passwords"

		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := Sign(&user); err != nil {
		resp["Register"] = "Error create user"

		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp["Register"] = "OK"
	c.JSON(http.StatusOK, resp)
}

func LoginUser(c *gin.Context) {
	resp := make(map[string]string)

	var user UserLogin
	bodyBytes, _ := io.ReadAll(c.Request.Body)

	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if Login(c.Writer, c.Request, &user) {
		resp["Login"] = "OK"
		c.JSON(http.StatusOK, resp)
	} else {
		fmt.Println("error login")

		resp["Login"] = "error login user"
		c.JSON(http.StatusForbidden, resp)
	}
}

func GetAllProducts(c *gin.Context) {
	//loginUser, user := CheckSesionUser(c.Request)

	// To Do: check if restouran

	//if !loginUser{
	//	c.Writer.WriteHeader(http.StatusUnauthorized)
	//	return
	//}

	var products []Product

	if err := DB.Find(&products).Error; err != nil {
		fmt.Println("error get user")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := make([]map[string]string, len(products))

	for i, product := range products {
		item := make(map[string]string)
		item["Id"] = strconv.FormatUint(product.Id, 10)
		item["Name"] = product.Name
		item["Image"] = product.Image
		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}
