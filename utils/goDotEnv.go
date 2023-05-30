package utils

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
)

func SetUpEnv() { //this is not for production environment only for dev, as we cannot load .ENV file in lambda

	err := godotenv.Load("../.env")

	if err != nil {
		log.Printf("Error loading env file %v", err)
		ServerError(errors.New("Error loading env file"))
	}
	log.Println("loaded env var")
}
