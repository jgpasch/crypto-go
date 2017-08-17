package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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

	if u.Number == "" || len(u.Number) < 10 {
		respondWithError(w, http.StatusBadRequest, "Please enter a valid phone number")
		return
	}

	defer r.Body.Close()

	u.Number = "1" + u.Number
	if err := u.createUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// try to send nexmo verifcation, if not fail for now
	body := map[string]string{"api_key": "8711bd32", "api_secret": "5ddcbb3b27977312", "number": u.Number, "brand": "NexmoVerifyTest"}
	jsonBody, _ := json.Marshal(body)

	res, err := http.Post("https://api.nexmo.com/verify/json", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue with nexmo, please try again")
		return
	}
	defer res.Body.Close()

	m := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error parsing")
		return
	}

	id, ok := m["request_id"].(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "error parsing")
		return
	}

	u.RequestID = id
	if err := u.updateUserRequestID(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue saving request ID")
		return
	}
	// fmt.Println("would be sending the verification code now and saving request id")
	// fmt.Println(u.Number)

	// respond with token
	tokenStr, err := createToken(u.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payloadStr := `{"email":` + `"` + u.Email + `"` + `,"token":` + `"` + tokenStr + `"}`
	payload := []byte(payloadStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(payload)
}

// --------------------
// end of createUser
// --------------------

// submit request ID to complete verification process
func (a *App) submitCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]
	fmt.Println(code)

	userEmail := context.Get(r, emailCtxKey)
	if userEmail == nil {
		respondWithError(w, http.StatusInternalServerError, "problem getting user email from token")
		return
	}

	var u user
	u.Email = userEmail.(string)
	if err := u.getUserByEmail(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting user by email")
		return
	}

	body := map[string]string{"api_key": "8711bd32", "api_secret": "5ddcbb3b27977312", "request_id": u.RequestID, "code": code}
	jsonBody, _ := json.Marshal(body)
	res, err := http.Post("https://api.nexmo.com/verify/check/json", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error sending verification code to nexmo")
		return
	}

	m := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error parsing")
		return
	}

	if m["status"] == "0" {
		sendMessage(a.nexmoClient, u.Number, "You have been verified!")
		respondWithJSON(w, http.StatusOK, m)
	} else {
		respondWithError(w, http.StatusBadRequest, m["error_text"].(string))
	}
	// if code == "1234" {
	// 	body := `{"data": "you did it!"}`
	// 	payload := []byte(body)
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusCreated)
	// 	w.Write(payload)
	// } else {
	// 	respondWithError(w, http.StatusBadRequest, "Wrong code!")
	// }

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

func (a *App) updateUserNumber(w http.ResponseWriter, r *http.Request) {
	userEmail := context.Get(r, emailCtxKey)
	if userEmail == nil {
		respondWithError(w, http.StatusInternalServerError, "Problem getting username from token")
		return
	}

	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	defer r.Body.Close()

	u.Email = userEmail.(string)

	if err := u.updateUserNumber(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) updateUserRequestID(w http.ResponseWriter, r *http.Request) {
	userEmail := context.Get(r, emailCtxKey)
	if userEmail == nil {
		respondWithError(w, http.StatusInternalServerError, "Problem getting username from token")
		return
	}

	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	defer r.Body.Close()

	u.Email = userEmail.(string)

	if err := u.updateUserRequestID(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}
