package pkg

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	Id       uint64 `gorm:"primaryKey"`
	Address  string
	IdBar    string
	Password string
	LngLatX  float64
	LngLatY  float64
}

func init() {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	DB = db

	if err != nil {
		panic("failed to connect database")
	}

	err = DB.AutoMigrate(&User{}, &Admin{}, &Product{}, &Bar{}, &TypeProduct{})
	if err != nil {
		panic("Error automigrate: ")
	}
}
