package helpers

import (
	"encoding/binary"
	"slices"
)

func CreateCNAMERecord(records map[string][]byte, fqdn string, domain string, ttl uint32) {

	name := convertNameToBytes(fqdn)

	aType := []byte{0x00, 0x01}

	qType := []byte{0x00, 0x05}

	// key for QDATA commands specifically looking for a CNAME
	keyCNAME := string(slices.Concat(name, qType))

	// key for general QDATA commands
	keyA := string(slices.Concat(name, aType))

	// we will always assume this request is coming through the internet
	class := []byte{0x00, 0x01}

	// set the TTL
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, ttl)

	// calculate rdata

	rdata := convertNameToBytes(domain)

	// encode the length of the ip address
	rdataLength := make([]byte, 2)
	binary.BigEndian.PutUint16(rdataLength, uint16(len(rdata)))

	response := slices.Concat(name, qType, class, ttlBytes, rdataLength, rdata)

	records[keyCNAME] = response
	records[keyA] = response
}
