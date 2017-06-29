package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"fmt"

	"log"

	"net/http"

	"encoding/json"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// App : Router + DB
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize Function to connect postgres driver
func (a *App) Initialize(user, dbname string) {
	connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=disable", user, dbname)
	// connectionString := fmt.Sprintf("user=john dbname=subscription_test sslmode=disable")
	fmt.Println(connectionString)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// Run runs the app
func (a *App) Run(addr string) {
	fmt.Printf("Server listening on port %s", strings.Trim(addr, ":"))
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	// user routes
	a.Router.HandleFunc("/user", a.getUserByEmail).Methods("POST")
	a.Router.HandleFunc("/users", a.getAllUsers).Methods("GET")
	a.Router.HandleFunc("/users/register", a.createUser).Methods("POST")
	a.Router.HandleFunc("/users/login", a.loginUser).Methods("POST")

	// subscription routes
	a.Router.HandleFunc("/subscriptions", a.getAllSubs).Methods("GET")
	a.Router.HandleFunc("/subscriptions", a.createSub).Methods("POST")
	a.Router.HandleFunc("/subscriptions/{token:[a-zA-Z]+}", a.getSubByToken).Methods("GET")
	a.Router.HandleFunc("/subscriptions/{id:[0-9]+}", a.getSub).Methods("GET")
	a.Router.HandleFunc("/subscriptions/{id:[0-9]+}", a.updateSub).Methods("PUT")
	a.Router.HandleFunc("/subscriptions/{id:[0-9]+}", a.deleteSub).Methods("DELETE")
}

// Create a token for the user
func createToken(username string) (string, error) {
	expiration := time.Now().Add(time.Hour * 1).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": username, "iat": expiration})

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

	respondWithJSON(w, http.StatusCreated, u)
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

// Gets a single subscription
func (a *App) getSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	s := sub{ID: id}
	if err := s.getSub(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Subscription not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

// GET sub by token
func (a *App) getSubByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	s := sub{Token: token}
	if err := s.getSubByToken(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Subscription not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

// GET all subscriptions
func (a *App) getAllSubs(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}

	if start < 0 {
		start = 0
	}

	subs, err := getAllSubs(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, subs)
}

func (a *App) createSub(w http.ResponseWriter, r *http.Request) {
	var s sub
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	defer r.Body.Close()

	if err := s.createSub(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, s)
}

func (a *App) updateSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var s sub
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	defer r.Body.Close()
	s.ID = id

	if err := s.updateSub(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) deleteSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Subscription ID")
		return
	}

	s := sub{ID: id}
	if err := s.deleteSub(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
