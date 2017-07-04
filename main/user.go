package main

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	Number    string `json:"number"`
	RequestID string `json:"request_id,omitempty"`
}

func (u *user) getUserByID(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (u *user) getUserByEmail(db *sql.DB) error {
	return db.QueryRow("SELECT email, password, number, request_id FROM users WHERE email=$1",
		u.Email).Scan(&u.Email, &u.Password, &u.Number, &u.RequestID)
}

func (u *user) createUser(db *sql.DB) error {
	// call get user by email to see if email already taken
	err := u.getUserByEmail(db)
	if err == nil {
		return errors.New("email is already in use")
	}

	// create hashed password using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = db.QueryRow(
		"INSERT INTO users(email, password) VALUES($1, $2) RETURNING id, email",
		u.Email, hash).Scan(&u.ID, &u.Email)

	u.Password = ""
	if err != nil {
		return err
	}

	return nil
}

func (u *user) comparePasswords(db *sql.DB) (int, error) {
	inputPassword := []byte(u.Password)
	u.Password = ""
	err := u.getUserByEmail(db)
	if err != nil {
		return 400, err
	}

	return 0, bcrypt.CompareHashAndPassword([]byte(u.Password), inputPassword)
}

func getAllUsers(db *sql.DB) ([]user, error) {
	rows, err := db.Query("SELECT id, email, number FROM users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []user{}

	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Email, &u.Number); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
