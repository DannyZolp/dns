package main

import (
	"log"
	"os"
	"sync"

	"github.com/DannyZolp/dns/management"
	"github.com/DannyZolp/dns/servers"
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
		wg.Add(3)

		go servers.UdpServer(records, &wg)

		go servers.TcpServer(records, &wg)

		go management.UpdateRecords(records, &wg)

		wg.Wait()
	} else {
		cli()
	}
}
