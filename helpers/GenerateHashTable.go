package helpers

import (
	"encoding/json"
	"os"
)

type A struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	TTL  int    `json:"ttl"`
}

type CNAME struct {
	Name   string `json:"name"`
	Target string `json:"target"`
	TTL    int    `json:"ttl"`
}

type MX struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Priority int    `json:"priority"`
	TTL      int    `json:"ttl"`
}

type TXT struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type records struct {
	A     []A     `json:"A"`
	CNAME []CNAME `json:"CNAME"`
	MX    []MX    `json:"MX"`
	TXT   []TXT   `json:"TXT"`
}

func GenerateHashTable() map[string][]byte {
	file, err := os.ReadFile("records.json")
	if err != nil {
		panic(err)
	}

	hashMap := make(map[string][]byte)

	parsed := records{}

	json.Unmarshal(file, &parsed)

	// create all A records
	for _, r := range parsed.A {
		createARecord(&hashMap, r.Name, r.IP, uint32(r.TTL))
	}

	// create CNAME records
	for _, r := range parsed.CNAME {
		createCNAMERecord(&hashMap, r.Name, r.Target, uint32(r.TTL))
	}

	// create MX records
	for _, r := range parsed.MX {
		createMXRecord(&hashMap, r.Name, r.Server, uint16(r.Priority), uint32(r.TTL))
	}

	// create TXT records
	for _, r := range parsed.TXT {
		createTXTRecord(&hashMap, r.Name, r.Content, uint32(r.TTL))
	}

	return hashMap
}
