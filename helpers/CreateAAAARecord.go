package helpers

import (
	"encoding/binary"
	"net"
	"slices"
)

func CreateAAAARecord(records map[string][]byte, fqdn string, ip string, ttl uint32) {
	ipAddr := net.ParseIP(ip)
	name := ConvertNameToBytes(fqdn)
	qType := []byte{0x00, 0x1C} // 28

	// key for map
	key := string(slices.Concat(name, qType))

	// we will always assume this request is coming through the internet
	class := []byte{0x00, 0x01}

	// set the TTL
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, ttl)

	// encode the length of the ip address
	ipLength := make([]byte, 2)
	binary.BigEndian.PutUint16(ipLength, uint16(len(ipAddr.To16())))

	response := slices.Concat(name, qType, class, ttlBytes, ipLength, []byte(ipAddr.To16()))

	records[key] = response
}
