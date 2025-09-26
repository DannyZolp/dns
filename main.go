package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	records := make(map[string][]byte)

	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Println("Starting udp handler...")
	go udp(records, &wg)

	fmt.Println("Starting tcp handler...")
	go tcp(records, &wg)

	fmt.Println("Starting management service...")
	go management(records, &wg)

	wg.Wait()
}
