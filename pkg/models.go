package pkg

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

type User struct {
	gorm.Model
	Id       uint64 `gorm:"primaryKey"`
	Username string
	Password string
	Email    string
	Phone    string
	Image    string
	Orders   []Order
}

type Order struct {
	gorm.Model
	Id            uint64 `gorm:"primaryKey"`
	UserID        uint64
	User          User `gorm:"foreignKey:UserID"`
	OrderProducts []OrderProduct
	BarID         uint64
	Bar           Bar `gorm:"foreignKey:BarID"`
	OrderTime     time.Time
	OrderId       uint64
}

type OrderProduct struct {
	gorm.Model
	OrderID   uint64
	Order     Order `gorm:"foreignKey:OrderID"`
	ProductID uint64
	Product   Product `gorm:"foreignKey:ProductID"`
	Count     uint8
	Status    string
}

type Admin struct {
	gorm.Model
	Id     uint64 `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserId"`
	UserId uint64
}

type Product struct {
	gorm.Model
	Id          uint64 `gorm:"primaryKey"`
	Image       string
	Name        string
	Type        TypeProduct `gorm:"foreignKey:TypeId"`
	TypeId      uint64
	Description string
}

type TypeProduct struct {
	gorm.Model
	Id   uint64 `gorm:"primaryKey"`
	Type string
}

type Bar struct {
	gorm.Model
	Id        uint64 `gorm:"primaryKey"`
	Address   string
	IdBar     string `gorm:"unique"`
	Password  string
	Longitude float64
	Latitude  float64
}

type ClientBarDB struct {
	gorm.Model
	RoomId string
	Bar    Bar `gorm:"foreignKey:BarId"`
	BarId  uint64
}

func init() {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	DB = db

	if err != nil {
		panic("failed to connect database")
	}

	err = DB.AutoMigrate(&User{}, &Admin{}, &Product{}, &Bar{}, &TypeProduct{}, Order{}, &OrderProduct{}, ClientBarDB{})
	if err != nil {
		panic("Error autoMigrate: ")
	}
}
