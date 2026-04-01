package main

import (
	"log"
	"os"
	"tcp-server/module/server"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file.")
	}
}

func main() {
	checker, err := server.GetJSONKeyChecker()
	if err != nil {
		log.Fatal("Cannot read key.json")
		return
	}

	app, err := server.CreateServer(os.Getenv("PORT"), checker)
	if err != nil {
		log.Fatal("Cannot create server: ", err)
		return
	}

	app.Listen()
	log.Println("Listening...")
	select {}
}
