package main

import (
	"database/sql"
)

type sub struct {
	ID           int     `json:"id"`
	Token        string  `json:"token"`
	Percent      float64 `json:"percent"`
	MinVal       float64 `json:"minVal"`
	MaxVal       float64 `json:"maxVal"`
	MinMaxChange float64 `json:"minMaxChange"`
	Active       bool    `json:"active"`
}

func (s *sub) getSubByToken(db *sql.DB) error {
	return db.QueryRow("SELECT token, percent, minval, maxval, minmaxchange, active FROM subs WHERE token=$1",
		s.Token).Scan(&s.Token, &s.Percent, &s.MinVal, &s.MaxVal, &s.MinMaxChange, &s.Active)
}

func (s *sub) getSub(db *sql.DB) error {
	return db.QueryRow("SELECT token, percent, minval, maxval, minmaxchange, active FROM subs WHERE id=$1",
		s.ID).Scan(&s.Token, &s.Percent, &s.MinVal, &s.MaxVal, &s.MinMaxChange, &s.Active)
}

func (s *sub) updateSub(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE subs SET token=$1, percent=$2, minval=$3, maxval=$4, minmaxchange=$5, active=$6 WHERE id=$7",
			s.Token, s.Percent, s.MinVal, s.MaxVal, s.MinMaxChange, s.Active, s.ID)

	return err
}

func (s *sub) deleteSub(db *sql.DB) error {
	_, err :=
		db.Exec("DELETE FROM subs WHERE id=$1", s.ID)

	return err
}

func (s *sub) createSub(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO subs(token, percent, minval, maxval, minmaxchange) VALUES($1, $2, $3, $4, $5) RETURNING id",
		s.Token, s.Percent, s.MinVal, s.MaxVal, s.MinMaxChange).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

func getAllSubs(db *sql.DB, start, count int) ([]sub, error) {
	rows, err := db.Query(
		"SELECT id, token, percent, minval, maxval, minmaxchange, active FROM subs LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	subs := []sub{}

	for rows.Next() {
		var s sub
		if err := rows.Scan(&s.ID, &s.Token, &s.Percent, &s.MinVal, &s.MaxVal, &s.MinMaxChange, &s.Active); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return subs, nil
}
