package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"slices"
	"sync"
)

func handleConnection(conn net.Conn, records map[string][]byte, headerFound []byte, headerNotFound []byte) {
	reader := bufio.NewReader(conn)

	lengthBytes := make([]byte, 2)
	_, err := reader.Read(lengthBytes)
	length := binary.BigEndian.Uint16(lengthBytes)

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
		fmt.Printf("Error reading content from TCP request  %v", err)
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

	record := records[string(request[12:endOfDomain+2])]

	if record != nil {
		response := slices.Concat(request[0:2], headerFound, request[12:endOfDomain+4], record)

		binary.BigEndian.PutUint16(lengthBytes, uint16(len(response)))

		conn.Write(slices.Concat(lengthBytes, response))
	} else {
		response := slices.Concat(request[0:2], headerNotFound, request[12:endOfDomain+4])

		binary.BigEndian.PutUint16(lengthBytes, uint16(len(response)))

		conn.Write(slices.Concat(lengthBytes, response))
	}

	conn.Close()
}

func tcp(port int, ip net.IP, records map[string][]byte, wg *sync.WaitGroup) {
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

	addr := net.TCPAddr{
		Port: port,
		IP:   ip,
	}

	ser, err := net.ListenTCP("tcp", &addr)

	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	fmt.Printf("Listening for TCP on %s:%d\n", ip.String(), port)

	for {
		conn, err := ser.Accept()

		if err != nil {
			fmt.Print("Error accepting TCP request")
			conn.Close()
			continue
		}

		go handleConnection(conn, records, headerFound, headerNotFound)
	}
}
