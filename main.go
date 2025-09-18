package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dnsIp := os.Getenv("DNS_SERVER_IP")
	dnsPort := os.Getenv("DNS_SERVER_PORT")

	mgmtIp := os.Getenv("MANAGEMENT_SERVER_IP")
	mgmtPortStr := os.Getenv("MANAGEMENT_SERVER_PORT")

	mgmtPort, _ := strconv.Atoi(mgmtPortStr)

	ip := net.ParseIP(dnsIp)
	port, _ := strconv.Atoi(dnsPort)

	records := make(map[string][]byte)

	var wg sync.WaitGroup
	wg.Add(3)

	fmt.Println("Starting udp handler...")
	go udp(port, ip, records, &wg)

	fmt.Println("Starting tcp handler...")
	go tcp(port, ip, records, &wg)

	fmt.Println("Starting management service...")
	go management(mgmtPort, mgmtIp, records, &wg)

	wg.Wait()
}
