package main

import (
	"foxy/internal/routes"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/", routes.GetImageHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
