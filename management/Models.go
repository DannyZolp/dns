package management

import "gorm.io/gorm"

type Zone struct {
	gorm.Model
	Name  string
	SOA   SOA
	NS    NS
	A     []A
	AAAA  []AAAA
	CNAME []CNAME
	MX    []MX
	TXT   []TXT
}

type SOA struct {
	gorm.Model
	SecondLevelDomain string
	SerialNumber      uint32
	TTL               uint32
	Refresh           uint32
	Retry             uint32
	Expire            uint32
	ZoneID            int
}

type NS struct {
	gorm.Model
	SecondLevelDomain string
	Nameservers       []string `gorm:"serializer:json"`
	TTL               uint32
	ZoneID            int
}

type A struct {
	gorm.Model
	Name   string
	IP     string
	TTL    uint32
	Zone   Zone
	ZoneID int
}

type AAAA struct {
	gorm.Model
	Name   string
	IP     string
	TTL    uint32
	Zone   Zone
	ZoneID int
}

type CNAME struct {
	gorm.Model
	Name   string
	Target string
	TTL    uint32
	Zone   Zone
	ZoneID int
}

type MX struct {
	gorm.Model
	Name     string
	Target   string
	Priority uint16
	TTL      uint32
	Zone     Zone
	ZoneID   int
}

type TXT struct {
	gorm.Model
	Name    string
	Content string
	TTL     uint32
	Zone    Zone
	ZoneID  int
}
