package servers

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/DannyZolp/dns/helpers"
)

func UdpServer(records map[string][]byte, wg *sync.WaitGroup) {
	defer wg.Done()

	packet := make([]byte, 512) // in udp DNS, the max length of a packet is 512 bytes [RFC1035]

	port, err := strconv.Atoi(os.Getenv("DNS_SERVER_PORT"))
	if err != nil {
		panic(err)
	}

	addr := net.UDPAddr{
		Port: port,
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

		record := records[string(packet[12:endOfDomain+2])]
		var response []byte

		if packet[endOfDomain+1 : endOfDomain+2][0] == 0x06 {
			// this is an SOA request
			fqdnParts := helpers.ConvertNameFromBytes(packet[12:endOfDomain])
			fqdn := strings.Join([]string{fqdnParts[len(fqdnParts)-2], fqdnParts[len(fqdnParts)-1]}, ".")

			response = slices.Concat(packet[0:2], HeaderFound, packet[12:endOfDomain+4], records[fqdn])
		} else if record != nil {
			response = slices.Concat(packet[0:2], HeaderFound, packet[12:endOfDomain+4], record)
		} else {
			response = slices.Concat(packet[0:2], HeaderNotFound, packet[12:endOfDomain+4])
		}

		ser.WriteToUDP(response, remoteAddr)
	}
}
