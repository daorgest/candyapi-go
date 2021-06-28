package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Candy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type candyHandlers struct {
	sync.Mutex
	store map[string]Candy
}

// handlers for Candy (e.g POST, GET, etc.)
func (h *candyHandlers) candies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}
}

// getting the candies from the GET request
func (h *candyHandlers) get(w http.ResponseWriter, r *http.Request) {
	candies := make([]Candy, len(h.store))
	h.Lock() // locking the access to the get request, we do this once
	i := 0
	for _, candy := range h.store { // for loope for getting everything stored
		candies[i] = candy
		i++
	}
	h.Unlock() // unlock the access to the request, we do this once
	jsonBytes, err := json.Marshal(candies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Add("content-type", "applicaiton/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// able to use status 302 to a random coaster
func (h *candyHandlers) getRandomCandy(w http.ResponseWriter, r *http.Request) {
	ids := make([]string, len(h.store)) // getting list of ids
	h.Lock()

	i := 0
	for id := range h.store { // iterating over the store
		ids[i] = id
		i++
	}
	defer h.Unlock() //done reading

	var target string
	if len(ids) == 0 { // if theres no candies
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 { // if thers 1 candy, search at index 0
		target = ids[0]
	} else { // if more than 1, than we can use a radom function
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))-1]
	}
	w.Header().Add("location", fmt.Sprintf("/coasters/%s", target))
	w.WriteHeader(http.StatusFound)
}

// this makes it able to get json data form the ID
func (h *candyHandlers) getCandy(w http.ResponseWriter, r *http.Request) {
	// finding whats after the slash
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if parts[2] == "random" {
		h.getRandomCandy(w, r)
		return
	}
	h.Lock() // locking the access to the get request, we do this once
	candy, ok := h.store[parts[2]]
	h.Unlock() // unlock the access to the request, we do this once
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(candy)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Add("content-type", "applicaiton/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// abling to post candies
func (h *candyHandlers) post(w http.ResponseWriter, r *http.Request) {
	// checking if the body is correct
	bodyBytes, err := ioutil.ReadAll(r.Body) //reading body
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	// checking to see if the header is not json
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType) // bad request error
		w.Write([]byte(fmt.Sprintf("Need content-type 'application/json', but got '%s'", ct)))
		return
	}

	var candy Candy
	err = json.Unmarshal(bodyBytes, &candy) // writing the inputted json to the candy array
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // if not entered in correcly
		w.Write([]byte(err.Error()))
		return
	}

	// generating a random id based on unix time in the form of base 10
	candy.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	h.Lock()
	h.store[candy.ID] = candy
	defer h.Unlock() // and breathe out :)
}

func newCandyHandlers() *candyHandlers {
	return &candyHandlers{
		store: map[string]Candy{},
	}
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		panic("required env var ADMIN_PASWORD not set")
	}
	return &adminPortal{password: password}
}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - unauthorized"))
		return
	}
	w.Write([]byte("<html><h1>Welcome to the admin portal</h1></html>"))
}

func main() {
	admin := newAdminPortal()
	candyHandlers := newCandyHandlers()
	http.HandleFunc("/candies", candyHandlers.candies)   //root for candies
	http.HandleFunc("/candies/", candyHandlers.getCandy) // another / for id's
	http.HandleFunc("/admin", admin.handler)             // admin pannel
	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		panic(err)
	}
}
