package main

import (
	"log"
	"sample/internal/api"
)

func main() {
	log.Println("application start!")
	api.StartServer()
	log.Println("application terminated!")
}
