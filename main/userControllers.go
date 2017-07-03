package main

import (
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Create a token for the user
func createToken(username string) (string, error) {
	timeNow := time.Now()
	expiration := time.Now().Add(time.Hour * 1).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": username, "iat": timeNow, "nbf": timeNow, "exp": expiration})

	tokenString, err := token.SignedString([]byte("mysecret"))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Loging a user in
func (a *App) loginUser(w http.ResponseWriter, r *http.Request) {
	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Payload Request")
		return
	}
	defer r.Body.Close()

	if statusCode, err := u.comparePasswords(a.DB); err != nil {
		switch statusCode {
		case 400:
			respondWithError(w, http.StatusBadRequest, "No User exists for this email")
			return
		case 0:
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// respond with token
	tokenStr, err := createToken(u.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payloadStr := `{"email":` + `"` + u.Email + `"` + `,"token":` + `"` + tokenStr + `"}`
	payload := []byte(payloadStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

// GET all users
func (a *App) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	// create user obj in postgres
	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	defer r.Body.Close()

	if err := u.createUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// respond with token
	tokenStr, err := createToken(u.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payloadStr := `{"email":` + `"` + u.Email + `"` + `,"token":` + `"` + tokenStr + `"}`
	payload := []byte(payloadStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

// GET user by email
func (a *App) getUserByEmail(w http.ResponseWriter, r *http.Request) {
	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	defer r.Body.Close()

	if err := u.getUserByEmail(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	u.Password = ""

	respondWithJSON(w, http.StatusOK, u)
}
