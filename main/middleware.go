package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

type contextKey int

const emailCtxKey contextKey = 0

// LoggingHandler function to log request info
func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

// ValidateToken middleware for validating token
func validateToken(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// get the jwt from the authorization header

		tokenStr := ""
		tokenStr = r.Header.Get("authorization")
		if tokenStr == "" {
			http.Error(w, "Please include a Token in authorization header", 400)
			return
		}
		// validate the jwt
		myToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("mysecret"), nil
		})

		if err != nil {
			http.Error(w, "Invalid Token", 401)
			return
		}

		if claims, ok := myToken.Claims.(jwt.MapClaims); ok && myToken.Valid {
			context.Set(r, emailCtxKey, claims["sub"])
			next.ServeHTTP(w, r)
		} else {
			fmt.Println(err)
		}
	}

	return http.HandlerFunc(fn)
}
