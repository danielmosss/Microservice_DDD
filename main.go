package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

	// TODO: You need to implement GET /persons/{id}, DELETE /persons/{id}, etc.

	fmt.Println("SocialGraphService running on :8080")
	http.ListenAndServe(":8080", mux)
}

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

func getPersons(w http.ResponseWriter, r *http.Request) {
	var _reqPersons []Person
	for _, p := range persons {
		_reqPersons = append(_reqPersons, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}

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
