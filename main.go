package main

import (
	"log"
	"tcp-server/module/server"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file.")
	}
}

type TestChecker struct{}

func (_ TestChecker) Check(name string, key string) bool {
	return true
}

func main() {
	checker := &TestChecker{}
	app, err := server.CreateServer(":3000", checker)
	if err != nil {
		log.Fatal("Cannot create server: ", err)
		return
	}
	app.Listen()
	log.Println("Listening...")
	select {}
}
