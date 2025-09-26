package main

import (
	"container/list"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"dannyzolp.com/m/v2/helpers"
)

func udp(records map[string][]byte, slds *list.List, wg *sync.WaitGroup) {
	defer wg.Done()

	packet := make([]byte, 512) // in udp DNS, the max length of a packet is 512 bytes [RFC1035]

	port, err := strconv.Atoi(os.Getenv("DNS_SERVER_PORT"))
	if err != nil {
		panic(err)
	}

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(os.Getenv("DNS_SERVER_IP")),
	}

	ser, err := net.ListenUDP("udp", &addr)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening for UDP on %s:%d\n", addr.IP.String(), port)

	for {
		_, remoteAddr, err := ser.ReadFromUDP(packet)

		if err != nil {
			fmt.Printf("UDP Error: %v", err)
			continue
		}

		endOfDomain := 12

		for {
			if packet[endOfDomain] == 0x00 {
				endOfDomain++
				break
			}
			endOfDomain++
		}

		if packet[endOfDomain+1 : endOfDomain+2][0] == 0x06 {
			// this is an SOA request
			domain := string(packet[12:endOfDomain])
			wroteResponse := false

			for e := slds.Front(); e != nil; e = e.Next() {
				if strings.Contains(domain, e.Value.(string)) {
					ser.WriteToUDP(slices.Concat(packet[0:2], HeaderFound, packet[12:endOfDomain+4], helpers.GenerateSOAResponse(e.Value.(string), getSerial(), 3600, 1800, 60, 3600)), remoteAddr)
					wroteResponse = true
					break
				}
			}

			if wroteResponse {
				continue
			}

		}

		record := records[string(packet[12:endOfDomain+2])]
		var response []byte

		if record != nil {
			response = slices.Concat(packet[0:2], HeaderFound, packet[12:endOfDomain+4], record)
		} else {
			response = slices.Concat(packet[0:2], HeaderNotFound, packet[12:endOfDomain+4])
		}

		ser.WriteToUDP(response, remoteAddr)
	}
}
