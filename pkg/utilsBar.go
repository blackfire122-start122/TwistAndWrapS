package pkg

import (
	. "TwistAndWrapS/pkg/logging"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
)

type BarLogin struct {
	IdBar    string
	Password string
}

func LoginB(w http.ResponseWriter, r *http.Request, barLogin *BarLogin) bool {
	session, _ := store.Get(r, "session-name")

	var bar Bar
	if err := DB.First(&bar, "id_bar = ?", barLogin.IdBar).Error; err != nil {
		ErrorLogger.Println(err.Error())
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(bar.Password), []byte(barLogin.Password))
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
	IdBar     string `json:"idBar"`
	Address   string `json:"address"`
	Password  string `json:"password"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

func SignBar(bar *BarRegister) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(bar.Password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	longitude, err := strconv.ParseFloat(bar.Longitude, 10)
	if err != nil {
		return "", err
	}

	latitude, err := strconv.ParseFloat(bar.Latitude, 10)
	if err != nil {
		return "", err
	}

	bar.IdBar = GenerateIdBar(longitude, latitude)

	DB.Create(&Bar{IdBar: bar.IdBar, Password: string(hashedPassword), Address: bar.Address, Longitude: longitude, Latitude: latitude})
	return bar.IdBar, err
}

func CheckSessionBar(r *http.Request) (bool, Bar) {
	session, _ := store.Get(r, "session-name")

	var bar Bar

	if session.IsNew {
		return false, bar
	}

	if err := DB.First(&bar, "id_bar = ?", session.Values["idBar"]).Error; err != nil {
		ErrorLogger.Println(err.Error())
		return false, bar
	}

	if session.Values["password"] != bar.Password {
		return false, bar
	}
	return true, bar
}

func GenerateIdBar(lngLatY float64, lngLatX float64) string {
	id := strconv.FormatFloat(lngLatY, 'f', -1, 64) + strconv.FormatFloat(lngLatX, 'f', -1, 64)
	id = strings.ReplaceAll(id, ".", "")
	id = strings.ReplaceAll(id, "-", "")
	return id
}
