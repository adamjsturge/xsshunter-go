package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretPublic, jwtSecretPrivate = generate_jwt_secret()

type Claims struct {
	payload string
	jwt.RegisteredClaims
}

func generate_jwt_secret() (ed25519.PublicKey, ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate ed25519 key pair: %v", err)
	}
	return publicKey, privateKey
}

func generate_and_set_jwt(w http.ResponseWriter, r *http.Request) {
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
		Path:    "/",
		Secure:  r.TLS != nil,
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
		return jwtSecretPublic, nil
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
	return token.SignedString(jwtSecretPrivate)
}
