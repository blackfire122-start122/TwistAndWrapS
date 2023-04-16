package pkg

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
)

type BarLogin struct {
	IdBar    string
	Password string
}

func LoginB(w http.ResponseWriter, r *http.Request, barLogin *BarLogin) bool {
	session, _ := store.Get(r, "session-name")

	var bar Bar
	err := DB.First(&bar, "id_bar = ?", barLogin.IdBar).Error

	if err != nil {
		fmt.Println("error db")
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(bar.Password), []byte(barLogin.Password))
	if err == nil {
		session.Values["idBar"] = bar.IdBar
		session.Values["password"] = bar.Password

		err = session.Save(r, w)
		if err != nil {
			return false
		}
	} else {
		return false
	}

	return true
}

type BarRegister struct {
	IdBar    string
	Password string
	LngLatX  string
	LngLatY  string
}

func SignBar(bar *BarRegister) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(bar.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	lngLatY, err := strconv.ParseFloat(bar.LngLatY, 10)
	if err != nil {
		return err
	}

	lngLatX, err := strconv.ParseFloat(bar.LngLatX, 10)
	if err != nil {
		return err
	}

	DB.Create(&Bar{IdBar: bar.IdBar, Password: string(hashedPassword), LngLatY: lngLatY, LngLatX: lngLatX})
	return err
}

func CheckSessionBar(r *http.Request) (bool, Bar) {
	session, _ := store.Get(r, "session-name")

	var bar Bar

	if session.IsNew {
		fmt.Println("not sessions")
		return false, bar
	}

	err := DB.First(&bar, "id_bar = ?", session.Values["idBar"]).Error

	if err != nil {
		fmt.Println("error db")
		return false, bar
	}

	if session.Values["password"] != bar.Password {
		return false, bar
	}
	return true, bar
}
