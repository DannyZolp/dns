package main

import (
	"fmt"
	"net"
	"slices"
	"sync"
)

func udp(port int, ip net.IP, records map[string][]byte, wg *sync.WaitGroup) {
	defer wg.Done()

	headerFound := []byte{
		// QR/OPCODE Section
		0x80, 0x00,
		// QDCOUNT (assumes only one request was made)
		0x00, 0x01,
		// ANCOUNT (there is only one answer in this generated response)
		0x00, 0x01,
		// NSCOUNT, we're not doing anything with nameservers here
		0x00, 0x00,
		// ARCOUNT, no additional records either.
		0x00, 0x00,
	}

	headerNotFound := []byte{
		// QR/OPCODE Section
		0x80, 0x00,
		// QDCOUNT (assumes only one request was made)
		0x00, 0x01,
		// ANCOUNT (there is only one answer in this generated response)
		0x00, 0x00,
		// NSCOUNT, we're not doing anything with nameservers here
		0x00, 0x00,
		// ARCOUNT, no additional records either.
		0x00, 0x00,
	}

	packet := make([]byte, 512) // in udp DNS, the max length of a packet is 512 bytes [RFC1035]

	addr := net.UDPAddr{
		Port: port,
		IP:   ip,
	}

	ser, err := net.ListenUDP("udp", &addr)

	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	fmt.Printf("Listening for UDP on %s:%d\n", ip.String(), port)

	for {
		_, remoteAddr, err := ser.ReadFromUDP(packet)

		if err != nil {
			fmt.Printf("Some error  %v", err)
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

		record := records[string(packet[12:endOfDomain+2])]

		if record != nil {
			response := slices.Concat(packet[0:2], headerFound, packet[12:endOfDomain+4], record)
			ser.WriteToUDP(response, remoteAddr)
		} else {
			response := slices.Concat(packet[0:2], headerNotFound, packet[12:endOfDomain+4])
			ser.WriteToUDP(response, remoteAddr)
		}

	}
}
