package main

import "testing"

// ###
// POST http://localhost:8080/persons
// Content-Type: application/json

// {
// "name": "Ally the Krokodil",
// "email": "ally@krokodil.com"
// }

// ###
// @id = f66d1d00-d102-4d66-8e37-f2e28c804a68
// GET http://localhost:8080/person/{{id}}

// ###
// POST http://localhost:8080/persons
// Content-Type: application/json

// {
// "name": "Glowy Submarine",
// "email": "glowy@krokodil.com"
// }

// ###
// GET http://localhost:8080/persons
// Content-Type: application/json

// ###
// POST http://localhost:8080/relationships
// Content-Type: application/json

// {
// "from_person_id": "bee97866-a5ed-4d92-a3e8-92b0888a140f",
// "to_person_id": "b2b07214-ee18-438d-8c5c-96c2ce49e521",
// "type": "Marriage"
// }

// ###
// GET http://localhost:8080/relationships
// Content-Type: application/json

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/persons", createPerson).Methods("POST")
	router.HandleFunc("/person/{id}", getPerson).Methods("GET")
	router.HandleFunc("/persons", getPersons).Methods("GET")
	return router
}

func HelperCreatePerson(name, email string) string {
	testPerson := Person{
		Name:  name,
		Email: email,
	}

	rr := makeRequest(createPerson, "POST", "/persons", testPerson)
	var response map[string]string

	json.NewDecoder(rr.Body).Decode(&response)

	return response["person_id"]
}

func makeRequest(handler http.HandlerFunc, method, path string, payload interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if method == "GET" {
		req, _ = http.NewRequest(method, path, nil)
	} else {
		body, _ := json.Marshal(payload)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func makeRequestWithRouter(method, path string, payload interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if method == "GET" {
		req, _ = http.NewRequest(method, path, nil)
	} else {
		body, _ := json.Marshal(payload)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	}

	router := setupRouter()
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func TestCreatePerson(t *testing.T) {
	testPerson := Person{
		Name:  "Ally the Krokodil",
		Email: "ally@krokodil.com",
	}

	rr := makeRequest(createPerson, "POST", "/persons", testPerson)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var response map[string]string
	json.NewDecoder(rr.Body).Decode(&response)

	if response["person_id"] == "" {
		t.Error("expected person_id in response")
	}
}

func TestGetPerson(t *testing.T) {
	personId := HelperCreatePerson("Glowy Submarine", "glowy@krokodil.com")

	rr := makeRequestWithRouter("GET", "/person/"+personId, nil)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response Person
	json.NewDecoder(rr.Body).Decode(&response)

	if response.ID == "" {
		t.Error("expected person in response")
	}
}
