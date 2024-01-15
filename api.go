package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
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

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/id/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleAccountWithID), s))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTranfer))

	log.Printf("JSON API server running on %s", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

// handle /account
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	case "GET":
		return s.handleGetAccount(w, r)

	case "POST":
		return s.handleCreateAccount(w, r)

	default:
		return fmt.Errorf("%s method not allowed", r.Method)
	}
}

// handle /account/id/{id}
func (s *APIServer) handleAccountWithID(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	case "GET":
		return s.handleGetAccountByID(w, r)

	case "DELETE":
		return s.handleDeleteAccount(w, r)

	default:
		return fmt.Errorf("%s method not allowed", r.Method)
	}

}

// GET /account
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {

	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

// GET /account/id/{id}
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {

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
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {

	defer r.Body.Close()

	CreateAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(CreateAccountReq); err != nil {
		return fmt.Errorf("invaild json format")
	}

	account := NewAccount(CreateAccountReq.FirstName, CreateAccountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return fmt.Errorf("creating account \n %v", err)
	}

	tokenString, err := createJWTToken(account)
	if err != nil {
		return err
	}

	// cookie := &http.Cookie{
	// 	Name:     "x-jwt-token",
	// 	Value:    tokenString,
	// 	Expires:  time.Now().Add(time.Hour * 24 * 30),
	// 	HttpOnly: true,
	// }
	// http.SetCookie(w, cookie)

	w.Header().Add("x-jwt-token", tokenString)

	return WriteJSON(w, http.StatusCreated, account)
}

// DELETE /account/id/{id}
func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	id, err := getID(r)
	if err != nil {
		return err
	}

	if err = s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusAccepted, map[string]int{"delete": id})
}

func (s *APIServer) handleTranfer(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	case "POST":
		defer r.Body.Close()

		transferResq := new(TramsferRequest)
		if err := json.NewDecoder(r.Body).Decode(transferResq); err != nil {
			return err
		}
		return WriteJSON(w, http.StatusAccepted, transferResq)

	default:
		return fmt.Errorf("%s method not allowed", r.Method)
	}

}

func createJWTToken(acc *Account) (string, error) {

	secret := []byte(os.Getenv("JWT_SECRET"))

	// Create the Claims
	claims := &jwt.RegisteredClaims{
		Issuer:    strconv.FormatInt(acc.AccoutnNumber, 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to create token")
	}
	return tokenStr, nil
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusUnauthorized, APIError{Error: "access denied"})
}

func withJWTAuth(handlerFunc http.HandlerFunc, s *APIServer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("x-jwt-token")

		// cookie, err := r.Cookie("x-jwt-token")
		// if err != nil {
		// 	permissionDenied(w)
		// 	return
		// }

		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}

		claimIssuer, err := token.Claims.GetIssuer()
		if err != nil {
			permissionDenied(w)
			return
		}

		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.store.GetAccountByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		if claimIssuer != strconv.FormatInt(account.AccoutnNumber, 10) {
			permissionDenied(w)
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

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
