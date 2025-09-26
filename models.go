package main

import "gorm.io/gorm"

type A struct {
	gorm.Model
	Name string `json:"name"`
	IP   string `json:"ip"`
	TTL  uint32 `json:"ttl"`
}

type AAAA struct {
	gorm.Model
	Name string `json:"name"`
	IP   string `json:"ip"`
	TTL  uint32 `json:"ttl"`
}

type CNAME struct {
	gorm.Model
	Name   string `json:"name"`
	Target string `json:"target"`
	TTL    uint32 `json:"ttl"`
}

type MX struct {
	gorm.Model
	Name     string `json:"name"`
	Target   string `json:"target"`
	Priority uint16 `json:"priority"`
	TTL      uint32 `json:"ttl"`
}

type TXT struct {
	gorm.Model
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     uint32 `json:"ttl"`
}
