package main

import (
	"container/list"
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
	slds := list.New()

	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Println("Starting udp handler...")
	go udp(records, slds, &wg)

	fmt.Println("Starting tcp handler...")
	go tcp(records, slds, &wg)

	fmt.Println("Starting management service...")
	go management(records, slds, &wg)

	wg.Wait()
}
