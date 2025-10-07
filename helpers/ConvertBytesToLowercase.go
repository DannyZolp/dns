package helpers

func ConvertBytesToLowercase(bytes []byte) []byte {
	i := 0
	for i < len(bytes) {
		length := int(bytes[i])
		if length == 0 {
			i++
			break
		}
		i++
		if i+length > len(bytes) {
			break // invalid length
		}
		for i <= length {
			if bytes[i] <= 90 && bytes[i] >= 65 {
				bytes[i] += 32
			}
			i++
		}
	}
	return bytes
}
