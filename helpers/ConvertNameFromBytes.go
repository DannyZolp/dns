package helpers

func ConvertNameFromBytes(data []byte) []string {
	var (
		labels []string
		offset int
	)
	for offset < len(data) {
		length := int(data[offset])
		if length == 0 {
			offset++
			break
		}
		offset++
		if offset+length > len(data) {
			break // invalid length
		}
		labels = append(labels, string(data[offset:offset+length]))
		offset += length
	}
	return labels
}
