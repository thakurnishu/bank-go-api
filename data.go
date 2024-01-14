package main

import (
	"math/rand"
	"time"
)

type TramsferRequest struct {
	ToAccountNumber int     `json:"toacountnumber"`
	Amount          float64 `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Account struct {
	ID            int       `json:"id"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	AccoutnNumber int64     `json:"accountNumber"`
	Balance       float64   `json:"balance"`
	CreatedAt     time.Time `json:"createdAt"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		// ID:            rand.Intn(10000),
		FirstName:     firstName,
		LastName:      lastName,
		AccoutnNumber: int64(rand.Intn(10000000)),
		Balance:       0,
		CreatedAt:     time.Now().UTC(),
	}
}
