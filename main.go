// @title Social Graph API
// @version 1.0
// @description A microservice for managing social graph relationships between persons
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @license.name Apache 2.0
// @host localhost:8080
// @basePath /
package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swaggo/http-swagger"
	_ "DDD_Microservice/docs"
	"net/http"
	"sync"
	"time"
)

type Person struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Email          string            `json:"email"`
	ProfilePicture string            `json:"profile_picture,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type Relationship struct {
	FromPersonID string            `json:"from_person_id"`
	ToPersonID   string            `json:"to_person_id"`
	Type         string            `json:"type"`
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

var (
	persons       = make(map[string]Person)
	relationships = []Relationship{}
	mu            sync.RWMutex
)

func publishEvent(eventType string, data interface{}) {
	eventData, _ := json.Marshal(data)
	// Voor nu log ik gewoon de events, maar moet even kijken hoe je dit kunt doen dat het echt een event pub wordt.
	fmt.Printf("[EVENT LOG] Type: %s | Payload: %s\n", eventType, string(eventData))
}

func main() {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("POST /persons", createPerson)
	mux.HandleFunc("GET /persons", getPersons)
	mux.HandleFunc("GET /person/{id}", getPerson)
	mux.HandleFunc("POST /relationships", createRelationship)
	mux.HandleFunc("GET /relationships", getReplationships)

	// Swagger UI endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// TODO: You need to implement GET /persons/{id}, DELETE /persons/{id}, etc.

	fmt.Println("Service running on :8080")
	fmt.Println("Swagger UI available at http://localhost:8080/swagger/")
	http.ListenAndServe(":8080", mux)
}

// createPerson godoc
// @Summary Create a new person
// @Description Create a new person in the social graph
// @Tags persons
// @Accept json
// @Produce json
// @Param person body Person true "Person object"
// @Success 201 {object} map[string]string{person_id=string}
// @Failure 400 {string} string "Invalid request body"
// @Router /persons [post]
func createPerson(w http.ResponseWriter, r *http.Request) {
	var p Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.ID = uuid.New().String()

	mu.Lock()
	persons[p.ID] = p
	mu.Unlock()

	// Raise Event
	publishEvent("PersonCreated", map[string]string{
		"event_type": "PersonCreated",
		"person_id":  p.ID,
		"name":       p.Name,
		"email":      p.Email,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"person_id": p.ID})
}

// createRelationship godoc
// @Summary Create a relationship between two persons
// @Description Create a new relationship in the social graph
// @Tags relationships
// @Accept json
// @Produce json
// @Param relationship body Relationship true "Relationship object"
// @Success 201 "Relationship created"
// @Failure 400 {string} string "Invalid request body"
// @Router /relationships [post]
func createRelationship(w http.ResponseWriter, r *http.Request) {
	var rel Relationship
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rel.CreatedAt = time.Now()

	mu.Lock()
	relationships = append(relationships, rel)
	mu.Unlock()

	// Raise Event
	publishEvent("RelationshipCreated", map[string]string{
		"event_type":     "RelationshipCreated",
		"from_person_id": rel.FromPersonID,
		"to_person_id":   rel.ToPersonID,
		"type":           rel.Type,
	})

	w.WriteHeader(http.StatusCreated)
}

// getPersons godoc
// @Summary Get all persons
// @Description Get a list of all persons in the social graph
// @Tags persons
// @Produce json
// @Success 200 {object} map[string]Person
// @Router /persons [get]
func getPersons(w http.ResponseWriter, r *http.Request) {
	var _reqPersons []Person
	for _, p := range persons {
		_reqPersons = append(_reqPersons, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}

// getPerson godoc
// @Summary Get a person by ID
// @Description Get a specific person from the social graph
// @Tags persons
// @Produce json
// @Param id path string true "Person ID"
// @Success 200 {object} Person
// @Failure 404 {string} string "Person not found"
// @Router /person/{id} [get]
func getPerson(w http.ResponseWriter, r *http.Request) {
	personId := mux.Vars(r)["id"]

	var _reqPerson Person
	for _, p := range persons {
		if p.ID == personId {
			_reqPerson = p
			break
		}
	}

	if _reqPerson.ID == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(_reqPerson)
}

type relationshipsStruct struct {
	Type      string    `json:"type"`
	Persons   []Person  `json:"persons"`
	CreatedAt time.Time `json:"created_at"`
}

// getReplationships godoc
// @Summary Get all relationships
// @Description Get a list of all relationships in the social graph
// @Tags relationships
// @Produce json
// @Success 200 {array} relationshipsStruct
// @Router /relationships [get]
func getReplationships(w http.ResponseWriter, r *http.Request) {
	var _reqRelationships []relationshipsStruct
	for _, p := range relationships {
		var person1 Person
		var person2 Person
		person1 = persons[p.FromPersonID]
		person2 = persons[p.ToPersonID]
		_reqRelationships = append(_reqRelationships, relationshipsStruct{
			Type: p.Type,
			Persons: []Person{
				person1,
				person2,
			},
			CreatedAt: p.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(_reqRelationships)
}
