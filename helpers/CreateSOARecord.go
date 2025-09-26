package helpers

import (
	"encoding/binary"
	"os"
	"slices"
)

func GenerateSOAResponse(sld string, serial []byte, ttl uint32, refresh uint32, retry uint32, expire uint32) []byte {
	mname := ConvertNameToBytes(os.Getenv("DNS_SERVER_NAME"))
	rname := ConvertNameToBytes(os.Getenv("DNS_SERVER_ADMIN_EMAIL"))

	qType := make([]byte, 2)
	binary.BigEndian.PutUint16(qType, 6)

	// we will always assume this request is coming through the internet
	class := []byte{0x00, 0x01}

	// get refresh, retry and expire and ttl
	brefresh, bretry, bexpire, bttl := make([]byte, 4), make([]byte, 4), make([]byte, 4), make([]byte, 4)
	binary.BigEndian.PutUint32(brefresh, refresh)
	binary.BigEndian.PutUint32(bretry, retry)
	binary.BigEndian.PutUint32(bexpire, expire)
	binary.BigEndian.PutUint32(bttl, ttl)

	// calculate rdata

	data := slices.Concat(mname, rname, serial, brefresh, bretry, bexpire, []byte{0x00, 0x00, 0x00, 0x00})

	// encode the length of the ip address
	rdataLength := make([]byte, 2)
	binary.BigEndian.PutUint16(rdataLength, uint16(len(data)))

	response := slices.Concat([]byte(sld), qType, class, bttl, rdataLength, data)

	return response
}
