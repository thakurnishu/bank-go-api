package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

type APIError struct {
	Error string `json:"error"`
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{
				Error: err.Error(),
			})
		}
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handAccount))
	router.HandleFunc("/account/id/{id}", makeHTTPHandleFunc(s.handAccountWithID))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handTranfer))

	log.Printf("JSON API server running on %s", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

// handle /account
func (s *APIServer) handAccount(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	case "GET":
		return s.handGetAccount(w, r)

	case "POST":
		return s.handCreateAccount(w, r)

	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

// handle /account/id/{id}
func (s *APIServer) handAccountWithID(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	case "GET":
		return s.handGetAccountByID(w, r)

	case "DELETE":
		return s.handDeleteAccount(w, r)

	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}

}

// GET /account
func (s *APIServer) handGetAccount(w http.ResponseWriter, r *http.Request) error {

	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

// GET /account/id/{id}
func (s *APIServer) handGetAccountByID(w http.ResponseWriter, r *http.Request) error {

	id, err := getID(r)
	if err != nil {
		return err
	}
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

// POST /account
func (s *APIServer) handCreateAccount(w http.ResponseWriter, r *http.Request) error {

	defer r.Body.Close()

	CreateAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(CreateAccountReq); err != nil {
		return fmt.Errorf("decoding account \n %v", err)
	}

	account := NewAccount(CreateAccountReq.FirstName, CreateAccountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return fmt.Errorf("creating account \n %v", err)
	}

	return WriteJSON(w, http.StatusCreated, account)
}

// DELETE /account/id/{id}
func (s *APIServer) handDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	id, err := getID(r)
	if err != nil {
		return err
	}

	if err = s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusAccepted, map[string]int{"delete": id})
}

func (s *APIServer) handTranfer(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	case "POST":
		defer r.Body.Close()

		transferResq := new(TramsferRequest)
		if err := json.NewDecoder(r.Body).Decode(transferResq); err != nil {
			return err
		}
		return WriteJSON(w, http.StatusAccepted, transferResq)

	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}

}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id %v", idStr)
	}
	return id, nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)

}
