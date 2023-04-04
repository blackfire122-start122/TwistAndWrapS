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
}

type Admin struct {
	gorm.Model
	Id     uint64 `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserId"`
	UserId uint64
}

type Product struct {
	gorm.Model
	Id    uint64 `gorm:"primaryKey"`
	Image string
	Name  string
}

func init() {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	DB = db

	if err != nil {
		panic("failed to connect database")
	}

	err = DB.AutoMigrate(&User{}, &Admin{}, &Product{})
	if err != nil {
		panic("Error automigrate: ")
	}
}
