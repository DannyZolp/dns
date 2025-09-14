package helpers

import (
	"encoding/binary"
	"slices"
)

func createMXRecord(recordsPointer *map[string][]byte, fqdn string, domain string, priority uint16, ttl uint32) {
	records := *recordsPointer

	name := convertNameToBytes(fqdn)

	qType := make([]byte, 2)
	binary.BigEndian.PutUint16(qType, 15)

	// key for QDATA commands specifically looking for a CNAME
	keyMX := string(slices.Concat(name, qType))

	// we will always assume this request is coming through the internet
	class := []byte{0x00, 0x01}

	// set the TTL
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, ttl)

	// calculate rdata

	priorityBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(priorityBytes, priority)

	rdata := slices.Concat(priorityBytes, convertNameToBytes(domain))

	// encode the length of the ip address
	rdataLength := make([]byte, 2)
	binary.BigEndian.PutUint16(rdataLength, uint16(len(rdata)))

	response := slices.Concat(name, qType, class, ttlBytes, rdataLength, rdata)

	records[keyMX] = response
}
