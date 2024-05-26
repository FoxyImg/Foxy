package main

import (
	"foxy/internal/env"
	"foxy/internal/routes"
	"net/http"
)

func main() {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}

	env.Boot()

	http.HandleFunc("/", routes.GetImageHandler)
	err := http.ListenAndServe(":"+env.FoxyEnvironment.Port, nil)
	if err != nil {
		panic(err)
	}
}
