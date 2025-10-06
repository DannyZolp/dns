package helpers

import (
	"encoding/binary"
	"fmt"
	"slices"
)

func CreateNSRecord(records map[string][]byte, fqdn string, nameservers []string, ttl uint32) {
	fmt.Println(len(nameservers))
	nameserverBytes := make([][]byte, len(nameservers))
	nsBytesLength := make([][]byte, len(nameservers))

	for i, nameserver := range nameservers {
		nameserverBytes[i] = ConvertNameToBytes(nameserver)
		nsBytesLength[i] = make([]byte, 2)
		binary.BigEndian.PutUint16(nsBytesLength[i], uint16(len(nameserverBytes[i])))
	}

	numberOfNameservers := uint16(len(nameservers))
	numOfNSBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(numOfNSBytes, numberOfNameservers)

	fmt.Printf("%x\n", numOfNSBytes)

	header := slices.Concat([]byte{0x80, 0x00, 0x00, 0x01}, numOfNSBytes, []byte{0x00, 0x00, 0x00, 0x00})

	fmt.Printf("% x, %d\n", header, len(header))

	name := ConvertNameToBytes(fqdn)
	qType := []byte{0x00, 0x02}

	// key for map
	key := string(slices.Concat(name, qType))

	// we will always assume this request is coming through the internet
	class := []byte{0x00, 0x01}

	// set the TTL
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, ttl)

	response := header

	for i, nsBytes := range nameserverBytes {
		response = slices.Concat(response, name, qType, class, ttlBytes, nsBytesLength[i], nsBytes)
	}

	fmt.Printf("% x\n", response)

	records[key] = response
}
