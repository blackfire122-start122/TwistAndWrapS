package internal

import (
	. "TwistAndWrapS/pkg"
	. "TwistAndWrapS/pkg/logging"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

	resp["id"] = strconv.FormatUint(user.Id, 10)
	resp["username"] = user.Username
	resp["email"] = user.Email
	resp["phone"] = user.Phone
	resp["image"] = user.Image
	resp["isAdmin"] = admin

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
		resp["Login"] = "error login user"
		c.JSON(http.StatusForbidden, resp)
	}
}

func LogoutUser(c *gin.Context) {
	resp := make(map[string]string)

	if Logout(c.Writer, c.Request) {
		resp["Logout"] = "OK"
		c.JSON(http.StatusOK, resp)
	} else {
		resp["Logout"] = "error logout user"
		c.JSON(http.StatusInternalServerError, resp)
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
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := make([]map[string]string, len(products))

	for i, product := range products {
		item := make(map[string]string)
		item["id"] = strconv.FormatUint(product.Id, 10)
		item["name"] = product.Name
		item["image"] = product.Image
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
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	resp := make(map[string]string)

	var bar BarRegister
	bodyBytes, _ := io.ReadAll(c.Request.Body)

	if err := json.Unmarshal(bodyBytes, &bar); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if bar.Password == "" || bar.Longitude == "" || bar.Latitude == "" || bar.Address == "" {
		resp["Register"] = "Not all field"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	idBar, err := SignBar(&bar)
	if err != nil {
		resp["Register"] = "Error create bar"

		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp["Register"] = "OK"
	resp["idBar"] = idBar
	c.JSON(http.StatusOK, resp)
}

func GetAllFoods(c *gin.Context) {
	var products []Product

	if err := DB.Preload("Type").Find(&products).Error; err != nil {
		ErrorLogger.Println("Error get products: " + err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make([]map[string]string, len(products))

	for i, product := range products {
		item := make(map[string]string)
		item["id"] = strconv.FormatUint(product.Id, 10)
		item["name"] = product.Name
		item["image"] = product.Image
		item["type"] = product.Type.Type
		item["description"] = product.Description
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
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var ImageName string

	if form.Image.Filename == "" && form.Image.Size == 0 {
		ImageName = user.Image
	} else {
		if err := c.SaveUploadedFile(form.Image, "./media/UserImages/"+user.Username+form.Image.Filename); err != nil {
			ErrorLogger.Println(err.Error())
		}
		if err := os.Remove("./" + user.Image); err != nil {
			ErrorLogger.Println(err.Error())
		}
		ImageName = "media/UserImages/" + user.Username + form.Image.Filename
	}

	if err := DB.Save(&User{Id: user.Id, Username: form.Username, Image: ImageName, Email: form.Email, Phone: form.Phone, Password: user.Password}).Error; err != nil {
		ErrorLogger.Println(err.Error())
	}

	if err := DB.First(&user, "id = ?", user.Id).Error; err != nil {
		ErrorLogger.Println(err.Error())
	}

	var admin string

	if CheckAdmin(user) {
		admin = "true"
	} else {
		admin = "false"
	}

	resp["id"] = strconv.FormatUint(user.Id, 10)
	resp["username"] = user.Username
	resp["email"] = user.Email
	resp["phone"] = user.Phone
	resp["image"] = user.Image
	resp["isAdmin"] = admin

	c.JSON(http.StatusOK, resp)
}

type FormCreateProduct struct {
	Type        string                `form:"type" binding:"required"`
	Name        string                `form:"name" binding:"required"`
	Description string                `form:"description" binding:"required"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
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
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var FileName string

	if form.File.Filename == "" && form.File.Size == 0 {
		FileName = ""
	} else {
		if err := c.SaveUploadedFile(form.File, "./media/ProductImages/"+form.Name+form.Type+form.File.Filename); err != nil {
			ErrorLogger.Println(err.Error())
		}
		FileName = "media/ProductImages/" + form.Name + form.Type + form.File.Filename
	}

	typeProductId, err := strconv.ParseUint(form.Type, 10, 64)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := DB.Create(&Product{Image: FileName, Name: form.Name, TypeId: typeProductId, Description: form.Description}).Error; err != nil {
		ErrorLogger.Println(err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
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
		ErrorLogger.Println(err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make([]map[string]string, len(types))

	for i, typeItem := range types {
		item := make(map[string]string)
		item["id"] = strconv.FormatUint(typeItem.Id, 10)
		item["type"] = typeItem.Type

		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}

func GetAllWorkedBars(c *gin.Context) {
	loginUser, _ := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var bars []Bar

	for bar, _ := range Clients {
		bars = append(bars, bar.Bar)
	}

	resp := make([]map[string]string, len(bars))

	for i, bar := range bars {
		item := make(map[string]string)
		item["id"] = strconv.FormatUint(bar.Id, 10)
		item["idBar"] = bar.IdBar
		item["address"] = bar.Address
		item["longitude"] = strconv.FormatFloat(bar.Longitude, 'f', -1, 64)
		item["latitude"] = strconv.FormatFloat(bar.Latitude, 'f', -1, 64)

		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}

type Food struct {
	Id    string `json:"id"`
	Count string `json:"count"`
}

type OrderForm struct {
	RestaurantId string `json:"restaurantId"`
	Foods        []Food `json:"foods"`
	Time         string `json:"time"`
}

type MsgToBarCreateOrder struct {
	FoodIdCount map[uint64]uint8
	Time        string
}

func OrderFood(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var form OrderForm
	if err := c.ShouldBind(&form); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	foodIdCount := make(map[uint64]uint8)
	var foodIds []uint64

	for _, food := range form.Foods {
		foodId, err := strconv.ParseUint(food.Id, 10, 64)

		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		foodIds = append(foodIds, foodId)

		foodCount, err := strconv.ParseUint(food.Count, 10, 8)

		if err != nil {
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

	msg, err := json.Marshal(MsgToBarCreateOrder{FoodIdCount: foodIdCount, Time: form.Time})
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
					orderTime, err := time.Parse("15:04", form.Time)
					if err != nil {
						c.Writer.WriteHeader(http.StatusBadRequest)
						return
					}

					var orderProducts []OrderProduct
					var products []Product
					DB.Find(&products, "id IN ?", foodIds)

					for _, product := range products {
						var status string
						if contains(m.ProductsCreated, product.Id) {
							status = "created"
						} else {
							status = "not created"
						}

						orderProduct := OrderProduct{
							Product: product,
							Count:   foodIdCount[product.Id],
							Status:  status,
						}
						orderProducts = append(orderProducts, orderProduct)
					}

					order := Order{
						User:          user,
						Bar:           bar,
						OrderProducts: orderProducts,
						OrderTime:     orderTime,
						OrderId:       m.Id,
					}

					if err := DB.Create(&order).Error; err != nil {
						c.Writer.WriteHeader(http.StatusInternalServerError)
						return
					}

					var orderProductsResp []respOrderOrderProducts
					for _, orderProduct := range order.OrderProducts {
						orderProductsResp = append(orderProductsResp, respOrderOrderProducts{
							Count:   orderProduct.Count,
							ID:      orderProduct.ID,
							OrderID: orderProduct.OrderID,
							Product: respOrderProduct{
								Description: orderProduct.Product.Description,
								Id:          orderProduct.Product.Id,
								Image:       orderProduct.Product.Image,
								Name:        orderProduct.Product.Name,
								Type:        orderProduct.Product.Type.Type,
							},
							Status: orderProduct.Status,
						})
					}

					c.JSON(http.StatusOK, respOrder{Id: order.Id, OrderId: order.OrderId, OrderProducts: orderProductsResp, OrderTime: order.OrderTime.Format("15:04")})
					return
				}
			}
		}
	}
}

func contains(slice []uint64, item uint64) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

type FormChangeFood struct {
	Id          string                `form:"id" binding:"required"`
	Type        string                `form:"type" binding:"required"`
	Name        string                `form:"name" binding:"required"`
	Description string                `form:"description" binding:"required"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
}

func ChangeFood(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	var form FormChangeFood
	if err := c.ShouldBind(&form); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var food Product

	if err := DB.First(&food, "id = ?", form.Id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	var FileName string

	if form.File.Filename == "" && form.File.Size == 0 {
		FileName = food.Image
	} else {
		if err := c.SaveUploadedFile(form.File, "./media/ProductImages/"+form.Name+form.Type+form.File.Filename); err != nil {
			ErrorLogger.Println(err.Error())
		}
		if err := os.Remove("./" + food.Image); err != nil {
			ErrorLogger.Println(err.Error())
		}
		FileName = "media/ProductImages/" + form.Name + form.Type + form.File.Filename
	}

	typeProductId, err := strconv.ParseUint(form.Type, 10, 64)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := DB.Save(&Product{Id: food.Id, Image: FileName, Name: form.Name, TypeId: typeProductId, Description: form.Description}).Error; err != nil {
		ErrorLogger.Println(err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := DB.Preload("Type").First(&food, "id = ?", food.Id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(map[string]string)
	resp["id"] = strconv.FormatUint(food.Id, 10)
	resp["image"] = food.Image
	resp["name"] = food.Name
	resp["description"] = food.Description
	resp["type"] = food.Type.Type

	c.JSON(http.StatusOK, resp)
}

func DeleteFood(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	id := c.Param("id")

	var food Product

	if err := DB.First(&food, "id = ?", id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	if err := DB.Delete(&Product{}, id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := os.Remove("./" + food.Image); err != nil {
		ErrorLogger.Println(err.Error())
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func GetAllBars(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	var bars []Bar

	if err := DB.Find(&bars).Error; err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}

	resp := make([]map[string]string, len(bars))

	for i, bar := range bars {
		item := make(map[string]string)
		item["id"] = strconv.FormatUint(bar.Id, 10)
		item["idBar"] = bar.IdBar
		item["address"] = bar.Address
		item["longitude"] = strconv.FormatFloat(bar.Longitude, 'f', -1, 64)
		item["latitude"] = strconv.FormatFloat(bar.Latitude, 'f', -1, 64)

		resp[i] = item
	}

	c.JSON(http.StatusOK, resp)
}

type FormChangeBar struct {
	Id        string `form:"id" binding:"required"`
	IdBar     string `form:"idBar" binding:"required"`
	Address   string `form:"address" binding:"required"`
	Longitude string `form:"longitude" binding:"required"`
	Latitude  string `form:"latitude" binding:"required"`
}

func ChangeBar(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	var form FormChangeBar
	if err := c.ShouldBind(&form); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var bar Bar

	if err := DB.First(&bar, "id = ?", form.Id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	longitude, err := strconv.ParseFloat(form.Longitude, 10)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	latitude, err := strconv.ParseFloat(form.Latitude, 10)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := DB.Save(&Bar{Id: bar.Id, IdBar: form.IdBar, Address: form.Address, Longitude: longitude, Latitude: latitude}).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: bars.id_bar") {
			c.Writer.WriteHeader(http.StatusConflict)
			return
		}
		ErrorLogger.Println(err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := DB.First(&bar, "id = ?", bar.Id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(map[string]string)
	resp["id"] = strconv.FormatUint(bar.Id, 10)
	resp["address"] = bar.Address
	resp["idBar"] = bar.IdBar
	resp["latitude"] = strconv.FormatFloat(bar.Latitude, 'f', -1, 64)
	resp["longitude"] = strconv.FormatFloat(bar.Longitude, 'f', -1, 64)

	c.JSON(http.StatusOK, resp)
}

func DeleteBar(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !CheckAdmin(user) {
		c.Writer.WriteHeader(http.StatusForbidden)
		return
	}

	id := c.Param("id")

	var bar Bar

	if err := DB.First(&bar, "id = ?", id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	if err := DB.Delete(&Bar{}, id).Error; err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

type respOrderProduct struct {
	Description string
	Id          uint64
	Image       string
	Name        string
	Type        string
}

type respOrderOrderProducts struct {
	Count   uint8
	ID      uint
	OrderID uint64
	Product respOrderProduct
	Status  string
}

type respOrder struct {
	Id            uint64
	OrderId       uint64
	OrderProducts []respOrderOrderProducts
	OrderTime     string
}

func GetOrders(c *gin.Context) {
	loginUser, user := CheckSessionUser(c.Request)

	if !loginUser {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	DB.Preload("Orders").Preload("Orders.OrderProducts.Product.Type").First(&user, "id = ?", user.Id)
	var resp []respOrder

	for _, order := range user.Orders {
		var orderProducts []respOrderOrderProducts
		for _, orderProduct := range order.OrderProducts {
			orderProducts = append(orderProducts, respOrderOrderProducts{
				Count:   orderProduct.Count,
				ID:      orderProduct.ID,
				OrderID: orderProduct.OrderID,
				Product: respOrderProduct{
					Description: orderProduct.Product.Description,
					Id:          orderProduct.Product.Id,
					Image:       orderProduct.Product.Image,
					Name:        orderProduct.Product.Name,
					Type:        orderProduct.Product.Type.Type,
				},
				Status: orderProduct.Status,
			})
		}

		resp = append(resp, respOrder{
			Id:            order.Id,
			OrderId:       order.OrderId,
			OrderProducts: orderProducts,
			OrderTime:     order.OrderTime.Format("15:04"),
		})
	}

	c.JSON(http.StatusOK, resp)
}
