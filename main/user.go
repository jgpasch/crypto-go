package main

import (
	"database/sql"
	"errors"
)

type user struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
}

func (u *user) getUserByID(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (u *user) getUserByEmail(db *sql.DB) error {
	return db.QueryRow("SELECT email, password, salt FROM users WHERE email=$1",
		u.Email).Scan(&u.Email, &u.Password, &u.Salt)
}

func (u *user) createUser(db *sql.DB) error {
	// call get user by email to see if email already taken
	err := u.getUserByEmail(db)
	if err == nil {
		return errors.New("email is already in use")
	}

	err = db.QueryRow(
		"INSERT INTO users(email, password, salt) VALUES($1, $2, $3) RETURNING id, email",
		u.Email, u.Password, "randomsalt").Scan(&u.ID, &u.Email)

	if err != nil {
		return err
	}

	return nil
	// if not, then generate a salt, save it
	// generate password with said salt, save it
	// and save the new user.
}

func getAllUsers(db *sql.DB) ([]user, error) {
	rows, err := db.Query("SELECT id, email, password, salt FROM users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []user{}

	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.Salt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
