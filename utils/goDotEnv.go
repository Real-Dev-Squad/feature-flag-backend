package utils

import (
	"errors"
	"log"

	"github.com/joho/godotenv"
)

func SetUpEnv() { //this is not for production environment only for dev, as we cannot load .ENV file in lambda

	err := godotenv.Load("../.env")

	if err != nil {
		log.Printf("Error loading env file \n %v", err)
		ServerError(errors.New("Error loading env file"))
	}
	log.Println("loaded env var")
}
