package utils

import (
	"github.com/joho/godotenv"
	"log"
	"errors"
)

func SetUpEnv() { //this is not for production environment only for dev, as we cannot load .ENV file in lambda

	err := godotenv.Load("../.env")

	if err != nil {
		log.Println("Error loading env file")
		ServerError(errors.New("Error loading env file"))
	}
	log.Println("loaded env var")
}
