package controller

import (
	"ciphertalk/common/constants"
	"ciphertalk/common/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var appSecret = []byte("super-secret-secret")
var tokenExpiration = time.Duration(24)

// User Database maps user name to id
var userDB = map[string]int{
	"foo": 1,
	"bar": 2,
}

// Login accepts client's request and generates a new JWT token for it
func (ctrl *APIController) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var loginReq = models.LoginRequest{}
	err := json.NewDecoder(r.Body).Decode(&loginReq)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if loginReq.UserName == "" {
		http.Error(w, "Invalid request. Missing user name", http.StatusBadRequest)
		return
	}

	response := models.LoginResponse{AuthToken: createToken(&loginReq.UserName)}
	payload, _ := json.Marshal(response)
	w.Header().Set(constants.HTTPContentType, constants.HTTPApplicationJSON)
	w.Write([]byte(payload))
}

func createToken(userName *string) string {
	// Create new auth token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set token claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = userName

	if userID, ok := userDB[*userName]; ok {
		claims["id"] = userID
	}

	claims["expires"] = time.Now().Add(time.Hour * tokenExpiration).Unix()

	// Sign the token with our secret
	tokenString, _ := token.SignedString(appSecret)

	return tokenString
}
