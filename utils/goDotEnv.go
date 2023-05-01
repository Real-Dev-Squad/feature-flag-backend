package utils

import (
	"github.com/joho/godotenv"
	"log"
)

func SetUpEnv() {

	err := godotenv.Load("../.env")

	if err != nil {
		log.Panic("Error loading .env file")
	}
	log.Println("loaded env var")
}
