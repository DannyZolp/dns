package helpers

func ConvertBytesToLowercase(bytes []byte) []byte {
	length := 0
	for i := 0; i < len(bytes); i++ {
		if length == 0 {
			length = int(bytes[i])
		} else {
			if bytes[i] >= 65 && bytes[i] <= 90 {
				bytes[i] += 32
			}
			length--
		}
	}
	return bytes
}
