package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"gopkg.in/njern/gonexmo.v1"
)

// App : Router + DB
type App struct {
	Router      *mux.Router
	DB          *sql.DB
	nexmoClient *nexmo.Client
}

// Initialize Function to connect postgres driver
func (a *App) Initialize(user, dbname string) {
	connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=disable", user, dbname)
	// connectionString := fmt.Sprintf("user=john dbname=subscription_test sslmode=disable")

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()

	// Test Routes
	// a.TestInitializeRoutes()
}

// Run runs the app
func (a *App) Run(addr string) {
	fmt.Printf("Server listening on port %s\n", strings.Trim(addr, ":"))

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	log.Fatal(http.ListenAndServe(addr, handlers.CORS(originsOk, headersOk)(a.Router)))

}

func (a *App) initializeRoutes() {
	// set up middleware using alice
	commonHandlers := alice.New(loggingHandler, validateToken)

	// user routes
	a.Router.Handle("/users", commonHandlers.ThenFunc(a.getUserByEmail)).Methods("POST")
	a.Router.Handle("/users", commonHandlers.ThenFunc(a.getAllUsers)).Methods("GET")
	a.Router.Handle("/users", commonHandlers.ThenFunc(a.updateUserNumber)).Methods("PUT")
	a.Router.Handle("/users/nexmo", commonHandlers.ThenFunc(a.updateUserRequestID)).Methods("PUT")

	a.Router.Handle("/auth/register", alice.New(loggingHandler).ThenFunc(a.createUser)).Methods("POST")
	a.Router.Handle("/auth/login", alice.New(loggingHandler).ThenFunc(a.loginUser)).Methods("POST")

	// subscription routes
	a.Router.Handle("/subscriptions", commonHandlers.ThenFunc(a.getAllSubs)).Methods("GET")
	a.Router.Handle("/subscriptions", commonHandlers.ThenFunc(a.createSub)).Methods("POST")
	a.Router.Handle("/subscriptions/{token:[a-zA-Z]+}", commonHandlers.ThenFunc(a.getSubByToken)).Methods("GET")
	a.Router.Handle("/subscriptions/{id:[0-9]+}", commonHandlers.ThenFunc(a.getSub)).Methods("GET")
	a.Router.Handle("/subscriptions/{id:[0-9]+}", commonHandlers.ThenFunc(a.updateSub)).Methods("PUT")
	a.Router.Handle("/subscriptions/{id:[0-9]+}", commonHandlers.ThenFunc(a.deleteSub)).Methods("DELETE")
}

// TestInitializeRoutes used to test routes without needing a token
func (a *App) TestInitializeRoutes() {

	a.Router.HandleFunc("/user", a.getUserByEmail).Methods("POST")
	a.Router.HandleFunc("/user", a.getAllUsers).Methods("POST")

	a.Router.HandleFunc("/auth/register", a.createUser).Methods("POST")
	a.Router.HandleFunc("/auth/login", a.loginUser).Methods("POST")

	// subscription routes
	a.Router.HandleFunc("/subscriptions", a.getAllSubs).Methods("GET")
	a.Router.HandleFunc("/subscriptions", a.createSub).Methods("POST")
	a.Router.HandleFunc("/subscriptions/{token:[a-zA-Z]+}", a.getSubByToken).Methods("GET")
	a.Router.HandleFunc("/subscriptions/{id:[0-9]+}", a.getSub).Methods("GET")
	a.Router.HandleFunc("/subscriptions/{id:[0-9]+}", a.updateSub).Methods("PUT")
	a.Router.HandleFunc("/subscriptions/{id:[0-9]+}", a.deleteSub).Methods("DELETE")
}
