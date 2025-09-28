package dns

import (
	"encoding/binary"
	"os"
	"strconv"
)

var HeaderFound = []byte{
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

var HeaderAuthorityFound = []byte{
	0x80, 0x00,
	0x00, 0x01,
	0x00, 0x00,
	0x00, 0x01,
	0x00, 0x00,
}

var HeaderNotFound = []byte{
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

func getSerial() []byte {
	serialBytes, _ := os.ReadFile(".serial")
	serialInt, _ := strconv.Atoi(string(serialBytes))
	serial := make([]byte, 4)
	binary.BigEndian.PutUint32(serial, uint32(serialInt))
	return serial
}
