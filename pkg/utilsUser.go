package pkg

import (
	"errors"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SECRET_KEY")))

type UserLogin struct {
	Username string
	Password string
}

func Login(w http.ResponseWriter, r *http.Request, userLogin *UserLogin) bool {
	session, _ := store.Get(r, "session-name")

	var user User
	if err := DB.First(&user, "Username = ?", userLogin.Username).Error; err != nil {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userLogin.Password))
	if err == nil {
		session.Values["id"] = user.Id
		session.Values["password"] = user.Password
		err = session.Save(r, w)
		if err != nil {
			return false
		}
	} else {
		return false
	}

	return true
}

type UserRegister struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func Sign(user *UserRegister) error {
	var users []User

	if err := DB.Where("Username = ?", user.Username).Find(&users).Error; err != nil {
		return err
	}
	if len(users) > 0 {
		return errors.New("user with the same username already exists")
	}

	if err := DB.Where("Email = ?", user.Email).Find(&users).Error; err != nil {
		return err
	}
	if len(users) > 0 {
		return errors.New("user with the same email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	DB.Create(&User{Username: user.Username, Password: string(hashedPassword), Email: user.Email})
	return err
}

func CheckSessionUser(r *http.Request) (bool, User) {
	session, _ := store.Get(r, "session-name")

	var user User

	if session.IsNew {
		return false, user
	}

	if err := DB.First(&user, "Id = ?", session.Values["id"]).Error; err != nil {
		return false, user
	}

	if session.Values["password"] != user.Password {
		return false, user
	}
	return true, user
}

func CheckAdmin(user User) bool {
	var admin Admin
	if err := DB.Where("user_id=?", user.Id).Find(&admin).Error; err != nil {
		return false
	}

	return admin.UserId == user.Id
}
