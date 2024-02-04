package main

import (
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = generate_jwt_secret()

type Claims struct {
	payload string
	jwt.RegisteredClaims
}

func generate_jwt_secret() string {
	newjwtSecret, err := get_secure_random_string(64)
	if err != nil {
		log.Fatal(err)
	}
	return newjwtSecret
}

func generate_and_set_jwt(w http.ResponseWriter) {
	expiration_time := time.Now().Add(24 * time.Hour)
	jwt, err := generate_jwt(expiration_time)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   jwt,
		Expires: expiration_time,
	})
}

func get_and_validate_jwt(r *http.Request) bool {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return false
	}

	return validate_jwt(cookie.Value)
}

func validate_jwt(token string) bool {
	parsed_token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return false
	}

	return parsed_token.Valid
}

func generate_jwt(expiration_time time.Time) (string, error) {
	time := time.Now()
	claims := &Claims{
		payload: time.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration_time),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(jwtSecret)
}
