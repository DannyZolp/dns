package main

import (
	"bufio"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"dannyzolp.com/m/v2/helpers"
)

func handleConnection(conn net.Conn, records map[string][]byte, slds *list.List) {
	reader := bufio.NewReader(conn)

	requestLength := make([]byte, 2)
	_, err := reader.Read(requestLength)
	length := binary.BigEndian.Uint16(requestLength)

	if length < 20 {
		conn.Close()
		return
	}

	if err != nil {
		fmt.Printf("Error reading length from TCP request  %v", err)
		conn.Close()
		return
	}

	request := make([]byte, length)
	_, err = reader.Read(request)

	if err != nil {
		if err != io.EOF {
			fmt.Printf("Error reading content from TCP request  %v", err)
		}
		conn.Close()
		return
	}

	endOfDomain := 12

	for {
		if request[endOfDomain] == 0x00 {
			endOfDomain++
			break
		}
		endOfDomain++
	}

	if request[endOfDomain+1 : endOfDomain+2][0] == 0x06 {
		// this is an SOA request
		domain := string(request[12:endOfDomain])

		for e := slds.Front(); e != nil; e = e.Next() {
			if strings.Contains(domain, e.Value.(string)) {
				res := slices.Concat(request[0:2], HeaderFound, request[12:endOfDomain+4], helpers.GenerateSOAResponse(e.Value.(string), getSerial(), 3600, 1800, 60, 3600))
				responseLength := make([]byte, 2)
				binary.BigEndian.PutUint16(responseLength, uint16(len(res)))
				conn.Write(slices.Concat(responseLength, res))
				conn.Close()

				return
			}
		}

	}

	record := records[string(request[12:endOfDomain+2])]

	var response []byte
	responseLength := make([]byte, 2)

	if record != nil {
		response = slices.Concat(request[0:2], HeaderFound, request[12:endOfDomain+4], record)
	} else {
		response = slices.Concat(request[0:2], HeaderNotFound, request[12:endOfDomain+4])
	}

	binary.BigEndian.PutUint16(responseLength, uint16(len(response)))

	conn.Write(slices.Concat(responseLength, response))

	conn.Close()
}

func tcp(records map[string][]byte, slds *list.List, wg *sync.WaitGroup) {
	defer wg.Done()

	port, err := strconv.Atoi(os.Getenv("DNS_SERVER_PORT"))
	if err != nil {
		panic(err)
	}

	addr := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP(os.Getenv("DNS_SERVER_IP")),
	}

	ser, err := net.ListenTCP("tcp", &addr)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening for TCP on %s:%d\n", addr.IP.String(), port)

	for {
		conn, err := ser.Accept()

		if err != nil {
			fmt.Printf("TCP Error: %v", err)
			conn.Close()
			continue
		}

		go handleConnection(conn, records, slds)
	}
}
