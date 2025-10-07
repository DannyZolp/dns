package servers

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/DannyZolp/dns/helpers"
)

func handleConnection(conn net.Conn, records map[string][]byte) {
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

	qType := request[endOfDomain : endOfDomain+2]

	record := records[string(helpers.ConvertBytesToLowercase(request[12:endOfDomain+2]))]

	var response []byte
	responseLength := make([]byte, 2)

	if qType[1] == 0x06 {
		// this is an SOA request
		fqdnParts := helpers.ConvertNameFromBytes(request[12:endOfDomain])
		fqdn := strings.Join([]string{fqdnParts[len(fqdnParts)-2], fqdnParts[len(fqdnParts)-1]}, ".")

		if records[fqdn] != nil {
			response = slices.Concat(request[0:2], HeaderFound, request[12:endOfDomain+4], records[fqdn])
		} else {
			response = slices.Concat(request[0:2], HeaderNotFound, request[12:endOfDomain+4])
		}
	} else if qType[1] == 0x02 {
		if len(record) > 10 {
			// this is an NS request
			response = slices.Concat(request[0:2], record[0:10], request[12:endOfDomain+4], record[10:])
		} else {
			response = slices.Concat(request[0:2], HeaderNotFound, request[12:endOfDomain+4])
		}
	} else if record != nil {
		response = slices.Concat(request[0:2], HeaderFound, request[12:endOfDomain+4], record)
	} else {
		response = slices.Concat(request[0:2], HeaderNotFound, request[12:endOfDomain+4])
	}

	binary.BigEndian.PutUint16(responseLength, uint16(len(response)))

	conn.Write(slices.Concat(responseLength, response))

	conn.Close()
}

func TcpServer(records map[string][]byte, wg *sync.WaitGroup) {
	defer wg.Done()

	port, err := strconv.Atoi(os.Getenv("DNS_SERVER_PORT"))
	if err != nil {
		panic(err)
	}

	addr := net.TCPAddr{
		IP:   net.ParseIP(os.Getenv("DNS_SERVER_IP")),
		Port: port,
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

		go handleConnection(conn, records)
	}
}
