package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/DannyZolp/dns/dns"
	"github.com/DannyZolp/dns/management"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) <= 1 {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		records := make(map[string][]byte)

		management.GenerateRecordMap(records)

		var wg sync.WaitGroup
		wg.Add(2)

		fmt.Println("Starting udp handler...")
		go dns.Udp(records, &wg)

		fmt.Println("Starting tcp handler...")
		go dns.Tcp(records, &wg)

		wg.Wait()
	} else {
		cli()
	}
}
