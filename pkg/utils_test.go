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
	// Підключення до тестової бази даних
	db, err := gorm.Open(sqlite.Open("test_database.db"), &gorm.Config{})
	DB = db

	// Створення таблиць в тестовій базі даних
	err = DB.AutoMigrate(&User{}, &Admin{}, &Product{}, &Bar{}, &TypeProduct{})

	// Додавання тестових даних до бази даних перед усіма тестами
	user := &UserRegister{
		Username: "testUser",
		Password: "testPassword",
		Email:    "test@example.com",
	}

	err = Sign(user)
	if err != nil {
		if err = DB.Where("Username = ?", user.Username).Delete(&User{}).Error; err != nil {
			return
		}
	}

	// Запуск усіх тестів
	code := m.Run()

	// Видалення тестових даних з бази даних після усіх тестів
	DB.Where("Username = ?", user.Username).Delete(&User{})

	// Вихід з тестової програми
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

	DB.Where("Username = ?", "testUser2").Delete(&User{})
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
