package main

import "gorm.io/gorm"

type SecondLevelDomain struct {
	gorm.Model
	Name   string `json:"name"`
	As     []A
	AAAAs  []AAAA
	CNAMEs []CNAME
	MXs    []MX
	TXTs   []TXT
}

type A struct {
	gorm.Model
	Name                string `json:"name"`
	IP                  string `json:"ip"`
	TTL                 uint32 `json:"ttl"`
	SecondLevelDomain   SecondLevelDomain
	SecondLevelDomainID int `json:"sldId"`
}

type AAAA struct {
	gorm.Model
	Name                string `json:"name"`
	IP                  string `json:"ip"`
	TTL                 uint32 `json:"ttl"`
	SecondLevelDomain   SecondLevelDomain
	SecondLevelDomainID int `json:"sldId"`
}

type CNAME struct {
	gorm.Model
	Name                string `json:"name"`
	Target              string `json:"target"`
	TTL                 uint32 `json:"ttl"`
	SecondLevelDomain   SecondLevelDomain
	SecondLevelDomainID int `json:"sldId"`
}

type MX struct {
	gorm.Model
	Name                string `json:"name"`
	Target              string `json:"target"`
	Priority            uint16 `json:"priority"`
	TTL                 uint32 `json:"ttl"`
	SecondLevelDomain   SecondLevelDomain
	SecondLevelDomainID int `json:"sldId"`
}

type TXT struct {
	gorm.Model
	Name                string `json:"name"`
	Content             string `json:"content"`
	TTL                 uint32 `json:"ttl"`
	SecondLevelDomain   SecondLevelDomain
	SecondLevelDomainID int `json:"sldId"`
}
