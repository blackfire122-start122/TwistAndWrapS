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

	if bar.Password == "" || bar.IdBar == "" || bar.LngLatX == "" || bar.LngLatY == "" {
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

	var form FormCreateProduct
	if err := c.ShouldBind(&form); err != nil {
		fmt.Println(err)
		return
	}

	var FileName string

	if form.File.Filename == "" && form.File.Size == 0 {
		FileName = ""
	} else {
		if err := c.SaveUploadedFile(form.File, "./media/ProductImages/"+form.Name+form.Type+form.File.Filename); err != nil {
			fmt.Println(err)
		}
		FileName = "media/ProductImages/" + form.Name + form.Type + form.File.Filename
	}

	typeProductId, err := strconv.ParseUint(form.Type, 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := DB.Create(&Product{Image: FileName, Name: form.Name, TypeId: typeProductId, Description: form.Description}).Error; err != nil {
		fmt.Println(err)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
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

func GetAllBars(c *gin.Context) {
	loginUser, _ := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var bars []Bar

	if err := DB.Find(&bars).Error; err != nil {
		fmt.Println("error get bars")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := make([]map[string]string, len(bars))

	for i, bar := range bars {
		item := make(map[string]string)
		item["Id"] = strconv.FormatUint(bar.Id, 10)
		item["IdBar"] = bar.IdBar
		item["Address"] = bar.Address
		item["LngLatX"] = strconv.FormatFloat(bar.LngLatX, 'f', -1, 64)
		item["LngLatY"] = strconv.FormatFloat(bar.LngLatY, 'f', -1, 64)

		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}

type Food struct {
	Id    string `json:"Id"`
	Count string `json:"Count"`
}

type Order struct {
	RestaurantId string `json:"RestaurantId"`
	Foods        []Food `json:"Foods"`
}

type respCreate struct {
	Type string
	Msg  string
}

func OrderFood(c *gin.Context) {
	loginUser, _ := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var form Order
	if err := c.ShouldBind(&form); err != nil {
		fmt.Println(err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	foodIdCount := make(map[uint64]uint8)
	for _, food := range form.Foods {
		foodId, err := strconv.ParseUint(food.Id, 10, 64)

		if err != nil {
			fmt.Println(err)
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		foodCount, err := strconv.ParseUint(food.Count, 10, 8)

		if err != nil {
			fmt.Println(err)
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if foodCount < 1 || foodCount > 10 {
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		foodIdCount[foodId] = uint8(foodCount)
	}

	var bar Bar
	if err := DB.First(&bar, "id_bar = ?", form.RestaurantId).Error; err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg, err := json.Marshal(foodIdCount)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	for cl, _ := range Clients {
		if cl.Bar.IdBar == bar.IdBar {
			Broadcast <- &Message{Type: "createOrder", Msg: string(msg), Client: cl}
			for {
				m := <-BroadcastReceiver
				if m.Client == cl {
					c.JSON(http.StatusOK, respCreate{Type: m.Type, Msg: m.Msg})
					break
				}
			}
		}
	}
}
