package internal

import (
	. "TwistAndWrapS/pkg"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

func GetUser(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var admin string

	if CheckAdmin(user) {
		admin = "true"
	} else {
		admin = "false"
	}

	resp := make(map[string]string)

	resp["Id"] = strconv.FormatUint(user.Id, 10)
	resp["Username"] = user.Username
	resp["Email"] = user.Email
	resp["Phone"] = user.Phone
	resp["Image"] = user.Image
	resp["IsAdmin"] = admin

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

	if user.Password == "" || user.Username == "" {
		resp["Register"] = "Not all field"

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
	loginBar, _ := CheckSessionBar(c.Request)

	if !loginBar {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

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

func LoginBar(c *gin.Context) {
	resp := make(map[string]string)

	var bar BarLogin
	bodyBytes, _ := io.ReadAll(c.Request.Body)

	if err := json.Unmarshal(bodyBytes, &bar); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if LoginB(c.Writer, c.Request, &bar) {
		resp["Login"] = "OK"
		c.JSON(http.StatusOK, resp)
	} else {
		resp["Login"] = "error login bar"
		c.JSON(http.StatusForbidden, resp)
	}
}

func RegisterBar(c *gin.Context) {
	resp := make(map[string]string)

	var bar BarRegister
	bodyBytes, _ := io.ReadAll(c.Request.Body)

	if err := json.Unmarshal(bodyBytes, &bar); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("here")

	if bar.Password == "" || bar.IdBar == "" {
		resp["Register"] = "Not all field"

		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := SignBar(&bar); err != nil {
		resp["Register"] = "Error create user"

		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp["Register"] = "OK"
	c.JSON(http.StatusOK, resp)
}

func GetAllFoods(c *gin.Context) {
	var products []Product

	if err := DB.Preload("Type").Find(&products).Error; err != nil {
		fmt.Println("error get products")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := make([]map[string]string, len(products))

	for i, product := range products {
		item := make(map[string]string)
		item["Id"] = strconv.FormatUint(product.Id, 10)
		item["Name"] = product.Name
		item["Image"] = product.Image
		item["Type"] = product.Type.Type
		item["Description"] = product.Description
		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}

type FormChangeUser struct {
	Username string                `form:"Username" binding:"required"`
	Email    string                `form:"Email" binding:"required"`
	Phone    string                `form:"Phone" binding:"required"`
	Image    *multipart.FileHeader `form:"Image" binding:"required"`
}

func ChangeUser(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp := make(map[string]string)
	var form FormChangeUser
	if err := c.ShouldBind(&form); err != nil {
		fmt.Println(err)
	}

	var ImageName string

	if form.Image.Filename == "" && form.Image.Size == 0 {
		ImageName = user.Image
	} else {
		if err := c.SaveUploadedFile(form.Image, "./media/UserImages/"+user.Username+form.Image.Filename); err != nil {
			fmt.Println(err)
		}
		if err := os.Remove("./" + user.Image); err != nil {
			fmt.Println(err)
		}
		ImageName = "media/UserImages/" + user.Username + form.Image.Filename
	}

	if err := DB.Save(&User{Id: user.Id, Username: form.Username, Image: ImageName, Email: form.Email, Phone: form.Phone, Password: user.Password}).Error; err != nil {
		fmt.Println(err)
	}

	if err := DB.First(&user, "id = ?", user.Id).Error; err != nil {
		fmt.Println(err)
	}

	resp["Id"] = strconv.FormatUint(user.Id, 10)
	resp["Username"] = user.Username
	resp["Email"] = user.Email
	resp["Phone"] = user.Phone
	resp["Image"] = user.Image

	c.JSON(http.StatusOK, resp)
}

type FormCreateProduct struct {
	Type        string                `form:"Type" binding:"required"`
	Name        string                `form:"Name" binding:"required"`
	Description string                `form:"Description" binding:"required"`
	File        *multipart.FileHeader `form:"File" binding:"required"`
}

func CreateProduct(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}
	fmt.Println("here")
	var form FormCreateProduct
	if err := c.ShouldBind(&form); err != nil {
		fmt.Println(err)
	}

	fmt.Println("here")

	var FileName string

	if form.File.Filename == "" && form.File.Size == 0 {
		FileName = ""
	} else {
		if err := c.SaveUploadedFile(form.File, "./media/ProductImages/"+form.Name+form.Type+form.File.Filename); err != nil {
			fmt.Println(err)
		}
		FileName = "media/ProductImages/" + form.Name + form.Type + form.File.Filename
	}
	fmt.Println("here")

	typeProductId, err := strconv.ParseUint(form.Type, 10, 64)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("here")

	if err := DB.Create(&Product{Image: FileName, Name: form.Name, TypeId: typeProductId, Description: form.Description}).Error; err != nil {
		fmt.Println(err)
	}
	fmt.Println("here")

}

func GetTypes(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	var types []TypeProduct

	if err := DB.Find(&types).Error; err != nil {
		fmt.Println("error get types")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := make([]map[string]string, len(types))

	for i, typeItem := range types {
		item := make(map[string]string)
		item["Id"] = strconv.FormatUint(typeItem.Id, 10)
		item["Type"] = typeItem.Type

		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}
