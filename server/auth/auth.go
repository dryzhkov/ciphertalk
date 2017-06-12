package auth

import (
	"errors"
	"fmt"
	"time"

	"strings"

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

// list of available chat clients. Map of username (string) to public key (32 bytes)
var registeredClients = make(map[string][32]byte)

func CreateToken(userName *string) string {
	// Create new auth token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set token claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = userName
	claims["expires"] = time.Now().Add(time.Hour * tokenExpiration).Unix()

	// Sign the token with our secret
	tokenString, _ := token.SignedString(appSecret)

	return tokenString
}

type UserProfile struct {
	AuthToken string
	UserName  string
}

func ParseToken(authHeader string) (UserProfile, error) {
	// assuming the auth header is in the format of "Bearer <token>", we only need the token value
	pieces := strings.Split(authHeader, " ")

	var userProfile = UserProfile{}
	if len(pieces) < 2 {
		return userProfile, errors.New("invalid authorization header")
	}
	tokenVal := pieces[1]
	token, err := jwt.Parse(tokenVal, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return appSecret, nil
	})

	if err != nil {
		return userProfile, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userProfile.AuthToken = tokenVal
		userProfile.UserName = claims["name"].(string)

		return userProfile, nil
	}

	return userProfile, err
}

func RegisterClient(userName string, pubKey [32]byte) {
	registeredClients[userName] = pubKey
}

func RetrieveClient(userName string) ([32]byte, error) {
	var res [32]byte
	if res, ok := registeredClients[userName]; ok {
		return res, nil
	}

	return res, errors.New("entry not found for key " + userName)
}
