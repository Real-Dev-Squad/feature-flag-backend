package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func SetUpEnv() {

	err := godotenv.Load("../.env")

	if err != nil {
		log.Panic("Error loading .env file")
	}
	log.Println("loaded env var")
}
