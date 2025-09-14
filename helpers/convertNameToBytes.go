package helpers

import (
	"strings"
)

func convertNameToBytes(fqdn string) []byte {
	bytes := make([]byte, len(fqdn)+2)

	splitFQDN := strings.Split(fqdn, ".")

	cursor := 0

	for _, part := range splitFQDN {
		bytes[cursor] = uint8(len(part))
		cursor++

		for _, char := range part {
			bytes[cursor] = byte(char)
			cursor++
		}
	}

	bytes[cursor] = 0x00

	return bytes
}
