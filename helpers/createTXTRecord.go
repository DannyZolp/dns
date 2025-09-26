package helpers

import (
	"encoding/binary"
	"slices"
)

func CreateTXTRecord(records map[string][]byte, fqdn string, data string, ttl uint32) {

	name := ConvertNameToBytes(fqdn)

	qType := make([]byte, 2)
	binary.BigEndian.PutUint16(qType, 16)

	// key for QDATA commands specifically looking for a CNAME
	keyTXT := string(slices.Concat(name, qType))

	// we will always assume this request is coming through the internet
	class := []byte{0x00, 0x01}

	// set the TTL
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, ttl)

	// calculate rdata

	rdata := slices.Concat([]byte{uint8(len(data))}, []byte(data))

	// encode the length of the ip address
	rdataLength := make([]byte, 2)
	binary.BigEndian.PutUint16(rdataLength, uint16(len(rdata)))

	response := slices.Concat(name, qType, class, ttlBytes, rdataLength, rdata)

	records[keyTXT] = response
}
