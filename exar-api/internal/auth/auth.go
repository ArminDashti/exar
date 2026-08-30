package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtSecret = "exar-backend-jwt-secret-change-in-production"

var users = map[string]user{
	"armin": {password: "Kp9#mX2vQwL4nT7", personID: 1},
	"ramin": {password: "Hn7$rT5yBcF8wJ3", personID: 2},
}

type user struct {
	password string
	personID int
}

type Claims struct {
	Username  string `json:"username"`
	PersonID  int    `json:"person_id"`
	jwt.RegisteredClaims
}

func Authenticate(username, password string) (string, int, error) {
	u, ok := users[username]
	if !ok || u.password != password {
		return "", 0, errors.New("invalid credentials")
	}

	token, err := issueToken(username, u.personID)
	if err != nil {
		return "", 0, err
	}

	return token, u.personID, nil
}

func issueToken(username string, personID int) (string, error) {
	claims := Claims{
		Username: username,
		PersonID: personID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
