package management

import (
	"context"

	"github.com/DannyZolp/dns/helpers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GenerateRecordMap(hashMap map[string][]byte) {
	db, err := gorm.Open(sqlite.Open("dns.db"), &gorm.Config{})
	ctx := context.Background()

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&A{}, &AAAA{}, &CNAME{}, &MX{}, &TXT{}, &SOA{})

	a, _ := gorm.G[A](db).Find(ctx)
	aaaa, _ := gorm.G[AAAA](db).Find(ctx)
	cname, _ := gorm.G[CNAME](db).Find(ctx)
	mx, _ := gorm.G[MX](db).Find(ctx)
	txt, _ := gorm.G[TXT](db).Find(ctx)
	soa, _ := gorm.G[SOA](db).Find(ctx)

	// create CNAME records
	for _, r := range cname {
		helpers.CreateCNAMERecord(hashMap, r.Name, r.Target, r.TTL)
	}

	// create MX records
	for _, r := range mx {
		helpers.CreateMXRecord(hashMap, r.Name, r.Target, r.Priority, r.TTL)
	}

	// create TXT records
	for _, r := range txt {
		helpers.CreateTXTRecord(hashMap, r.Name, r.Content, r.TTL)
	}

	for _, r := range aaaa {
		helpers.CreateAAAARecord(hashMap, r.Name, r.IP, r.TTL)
	}

	// create all A records
	for _, r := range a {
		helpers.CreateARecord(hashMap, r.Name, r.IP, r.TTL)
	}

	// create SOAs
	for _, r := range soa {
		helpers.CreateSOARecord(hashMap, r.SecondLevelDomain, r.SerialNumber, r.TTL, r.Refresh, r.Retry, r.Expire)
	}
}
