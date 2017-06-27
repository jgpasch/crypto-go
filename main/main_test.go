package main_test

import (
	"bytes"
	"crypto-go/main"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize("john", "new_sub_test_db")

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

// Test Empty Table
func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/subscriptions", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

// Test non existing subscription
func TestGetNonExistentSub(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/subscriptions/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Subscription not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Subscription not found'. Got '%s'", m["error"])
	}
}

// Test creating a subscription
func TestCreateSub(t *testing.T) {
	clearTable()

	payload := []byte(`{"token":"ETH","percent":10,"minVal":220,"maxVal":400,"minMaxChange":10}`)

	req, _ := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["token"] != "ETH" {
		t.Errorf("Expected the 'token' key of the response to be set to 'ETH'. Got '%s'", m["token"])
	}

	if m["percent"] != 10.0 {
		t.Errorf("Expected the 'percent' key of the response to be set to '10'. Got '%v'", m["percent"])
	}

	if m["minVal"] != 220.0 {
		t.Errorf("Expected the 'minval' key of the response to be set to '220'. Got '%v'", m["minVal"])
	}

	if m["maxVal"] != 400.0 {
		t.Errorf("Expected the 'maxval' key of the response to be set to '400'. Got '%v'", m["maxval"])
	}

	if m["minMaxChange"] != 10.0 {
		t.Errorf("Expected the 'minmaxchange' key of the response to be set to '10'. Got '%v'", m["minmaxchange"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

// Test to get single sub
func TestGetSub(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/subscriptions/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO subs(token, percent, minval, maxval, minmaxchange) VALUES($1, $2, $3, $4, $5)", "Sub "+strconv.Itoa(i), 10, (i+1.0)*20, (i+1.0)*30, 10)
	}
}

// Test to update a sub
func TestUpdateSub(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/subscriptions/1", nil)
	response := executeRequest(req)

	var originalSub map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalSub)

	payload := []byte(`{"token":"ETH - updated","percent":100,"minval":5,"maxval":3,"minmaxchange":65}`)

	req, _ = http.NewRequest("PUT", "/subscriptions/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	fmt.Println(m)
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["id"] != originalSub["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalSub["id"], m["id"])
	}

	if m["token"] == originalSub["token"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalSub["name"], m["name"], m["name"])
	}

	if m["percent"] == originalSub["percent"] {
		t.Errorf("Expected the percent to change from '%v' to '%v'. Got '%v'", originalSub["percent"], m["percent"], m["percent"])
	}

	if m["minVal"] == originalSub["minVal"] {
		t.Errorf("Expected the minVal to change from '%v' to '%v'. Got '%v'", originalSub["minVal"], m["minVal"], m["minVal"])
	}

	if m["maxVal"] == originalSub["maxVal"] {
		t.Errorf("Expected the maxVal to change from '%v' to '%v'. Got '%v'", originalSub["maxVal"], m["maxVal"], m["maxVal"])
	}

	if m["minMaxChange"] == originalSub["minMaxChange"] {
		t.Errorf("Expected the minMaxChange to change from '%v' to '%v'. Got '%v'", originalSub["minMaxChange"], m["minMaxChange"], m["minMaxChange"])
	}
}

// Test to delete a sub
func TestDeleteSub(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/subscriptions/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/subscriptions/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/subscriptions/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS subs
(
id SERIAL PRIMARY KEY,
token VARCHAR(30) NOT NULL,
percent NUMERIC(10,2) NOT NULL DEFAULT 0,
minval NUMERIC(10,2) NOT NULL DEFAULT 0,
maxval NUMERIC(10,2) NOT NULL DEFAULT 0,
minmaxchange NUMERIC(10,2) NOT NULL DEFAULT 0,
active BOOLEAN DEFAULT FALSE
)`

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE from subs")
	a.DB.Exec("ALTER SEQUENCE subs_id_seq RESTART WITH 1")
}
