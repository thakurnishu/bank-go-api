package main

import (
	"log"
)

func main() {

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatalf("Creating New Store \n %s\n", err.Error())
	}

	if err = store.init(); err != nil {
		log.Fatalf("Intilizing Database \n %s\n", err.Error())
	}

	server := NewAPIServer(":3000", store)
	server.Run()
}
