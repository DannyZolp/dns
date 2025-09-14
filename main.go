package main

import (
	"fmt"
	"net"
	"sync"

	"dannyzolp.com/m/v2/helpers"
)

func main() {
	ip, port := net.ParseIP("127.0.0.1"), 1053

	records := helpers.GenerateHashTable()

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Starting udp handler...")
	go udp(port, ip, records, &wg)

	fmt.Println("Starting tcp handler...")
	go tcp(port, ip, records, &wg)

	wg.Wait()
}
