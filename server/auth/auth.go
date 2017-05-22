package auth

import (
	"fmt"
	"time"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

var appSecret = []byte("super-secret-secret")
var tokenExpiration = time.Duration(24)

var JwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return appSecret, nil
	},
	// When set, the middleware verifies that tokens are signed with the specific signing algorithm
	// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
	// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
	SigningMethod: jwt.SigningMethodHS256,
})

// User Database maps user name to id
var userDB = map[string]int{
	"foo": 1,
	"bar": 2,
}

func CreateToken(userName *string) string {
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

type UserProfile struct {
	AuthToken string
	UserName  string
	UserID    string
}

func parseToken(tokenVal string) UserProfile {
	token, err := jwt.Parse(tokenVal, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return appSecret, nil
	})

	var userProfile = UserProfile{}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["foo"], claims["nbf"])
	} else {
		fmt.Println(err)
	}

	return userProfile
}
