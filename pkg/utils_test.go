package pkg

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	db, err := gorm.Open(sqlite.Open("test_database.db"), &gorm.Config{})
	DB = db

	err = DB.AutoMigrate(&User{}, &Admin{}, &Product{}, &Bar{}, &TypeProduct{})

	// Додавання тестових даних до бази даних перед усіма тестами
	user := &UserRegister{
		Username: "testUser",
		Password: "testPassword",
		Email:    "test@example.com",
	}

	bar := &BarRegister{
		Address:   "Test Address",
		Password:  "testPassword",
		Longitude: "1.234567",
		Latitude:  "2.345678",
	}

	err = Sign(user)
	if err != nil {
		if err = DB.Where("Username = ?", user.Username).Delete(&User{}).Error; err != nil {
			return
		}
	}

	_, err = SignBar(bar)
	if err != nil {
		if err = DB.Where("Address = ?", bar.Address).Delete(&Bar{}).Error; err != nil {
			return
		}
	}

	code := m.Run()

	// Видалення тестових даних з бази даних після усіх тестів
	DB.Unscoped().Where("Username = ?", user.Username).Delete(&User{})
	DB.Unscoped().Where("Address = ?", bar.Address).Delete(&Bar{})

	os.Exit(code)
}

func TestSign(t *testing.T) {
	// Створення тестового користувача
	user := &UserRegister{
		Username: "testUser2",
		Password: "testPassword2",
		Email:    "test@example2.com",
	}

	// Act
	err := Sign(user)

	// Assert
	assert.NoError(t, err)

	// Act: Повторна спроба створити користувача з тим самим іменем
	err = Sign(user)

	// Assert: Повинна повернутись помилка, що користувач вже існує
	assert.Error(t, err)
	assert.Equal(t, "user with the same username already exists", err.Error())

	// Act: Повторна спроба створити користувача з тією самою електронною адресою
	user = &UserRegister{
		Username: "anotherUser",
		Password: "anotherPassword",
		Email:    "test@example2.com",
	}
	err = Sign(user)

	// Assert: Повинна повернутись помилка, що користувач вже існує
	assert.Error(t, err)
	assert.Equal(t, "user with the same email already exists", err.Error())

	DB.Unscoped().Where("Username = ?", "testUser2").Delete(&User{})
	DB.Unscoped().Where("Username = ?", "anotherUser").Delete(&User{})
}

func TestLogin_Success(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "api/user/login", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	userLogin := &UserLogin{
		Username: "testUser",
		Password: "testPassword",
	}

	// Act
	result := Login(w, req, userLogin)

	// Assert
	assert.True(t, result)
}

func TestLogin_Failure(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "api/user/login", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	userLogin := &UserLogin{
		Username: "invalidUser",
		Password: "invalidPassword",
	}

	// Act
	result := Login(w, req, userLogin)

	// Assert
	assert.False(t, result)
}

func TestLoginB_Success(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/api/bar/loginBar", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	var bar Bar

	DB.First(&bar, "Address = ?", "Test Address")

	barLogin := &BarLogin{
		IdBar:    bar.IdBar,
		Password: "testPassword",
	}

	// Act
	result := LoginB(w, req, barLogin)

	// Assert
	assert.True(t, result)
}

func TestLoginB_Failure(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/api/bar/loginBar", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	barLogin := &BarLogin{
		IdBar:    "invalidBar",
		Password: "invalidPassword",
	}

	// Act
	result := LoginB(w, req, barLogin)

	// Assert
	assert.False(t, result)
}

func TestSignBar(t *testing.T) {
	// Arrange
	bar := &BarRegister{
		IdBar:     "",
		Address:   "Test Address 2",
		Password:  "testPassword",
		Longitude: "10.123456",
		Latitude:  "20.654321",
	}

	// Act
	id, err := SignBar(bar)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Act: Повторна спроба створити бар з тією самою адресою
	id, err = SignBar(bar)

	// Assert: Повинна повернутись помилка, що бар з такою адресою вже існує
	assert.Error(t, err)
	assert.Equal(t, "bar with the same address already exists", err.Error())

	// Act: Повторна спроба створити бар з тими самими координатами
	bar = &BarRegister{
		IdBar:     "",
		Address:   "Another Address",
		Password:  "testPassword",
		Longitude: "10.123456",
		Latitude:  "20.654321",
	}
	id, err = SignBar(bar)

	// Assert: Повинна повернутись помилка, що бар з такими координатами вже існує
	assert.Error(t, err)
	assert.Equal(t, "bar with the same longitude and latitude already exists", err.Error())

	DB.Unscoped().Where("Address = ?", "Test Address 2").Delete(&Bar{})
	DB.Unscoped().Where("Address = ?", "Another Address").Delete(&Bar{})
}
