package main

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *user) getUserByID(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (u *user) getUserByEmail(db *sql.DB) error {
	return db.QueryRow("SELECT email, password FROM users WHERE email=$1",
		u.Email).Scan(&u.Email, &u.Password)
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

	if err != nil {
		return err
	}

	return nil
	// and save the new user.
}

func (u *user) comparePasswords(db *sql.DB) (int, error) {
	inputPassword := []byte(u.Password)
	err := u.getUserByEmail(db)
	if err != nil {
		return 400, err
	}

	return 0, bcrypt.CompareHashAndPassword([]byte(u.Password), inputPassword)
}

func getAllUsers(db *sql.DB) ([]user, error) {
	rows, err := db.Query("SELECT id, email, password FROM users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []user{}

	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Email, &u.Password); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
