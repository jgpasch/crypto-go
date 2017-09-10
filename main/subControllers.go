package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// *****
// Gets a single subscription
func (a *App) getSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	s := sub{ID: id}
	userEmail := context.Get(r, emailCtxKey)
	s.Owner = userEmail.(string)
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

// *****
// GET sub by token
func (a *App) getSubByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	s := sub{Token: token}
	userEmail := context.Get(r, emailCtxKey)
	s.Owner = userEmail.(string)
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

// *****
// GET all subscriptions
func (a *App) getAllSubs(w http.ResponseWriter, r *http.Request) {
	userEmail := context.Get(r, emailCtxKey)
	owner := userEmail.(string)

	subs, err := getAllSubsByOwner(a.DB, owner)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, subs)
}

// *****
// *****
func (a *App) createSub(w http.ResponseWriter, r *http.Request) {
	var s sub
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	userEmail := context.Get(r, emailCtxKey)
	s.Owner = userEmail.(string)
	defer r.Body.Close()

	if err := s.createSub(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, s)
}

// *****
// *****
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

// *****
// *****
func (a *App) toggleActive(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	var s sub
	s.ID = id
	userEmail := context.Get(r, emailCtxKey)
	s.Owner = userEmail.(string)

	if err := s.toggleActive(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if s.Active {
		time.AfterFunc(1*time.Second, s.doEvery)
	} else {
		s.stopWatch()
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"result": "success", "active": s.Active})
}

// *****
// *****
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
