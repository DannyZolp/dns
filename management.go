package main

import (
	"container/list"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"dannyzolp.com/m/v2/helpers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func incrementSerial() {
	serialBytes, err := os.ReadFile(".serial")
	if err != nil {
		log.Default().Println(err)
	}
	serial, _ := strconv.Atoi(string(serialBytes))
	serial++

	os.WriteFile(".serial", []byte(strconv.Itoa(serial)), 0666)
}

func getSerial() []byte {
	serialBytes, _ := os.ReadFile(".serial")
	serialInt, _ := strconv.Atoi(string(serialBytes))
	serial := make([]byte, 4)
	binary.BigEndian.PutUint32(serial, uint32(serialInt))
	return serial
}

func generateHashTable(hashMap map[string][]byte, db *gorm.DB, ctx context.Context) {
	a, _ := gorm.G[A](db).Find(ctx)
	aaaa, _ := gorm.G[AAAA](db).Find(ctx)
	cname, _ := gorm.G[CNAME](db).Find(ctx)
	mx, _ := gorm.G[MX](db).Find(ctx)
	txt, _ := gorm.G[TXT](db).Find(ctx)

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
}

func management(records map[string][]byte, slds *list.List, wg *sync.WaitGroup) {
	defer wg.Done()

	mgmtIp := os.Getenv("MANAGEMENT_SERVER_IP")
	mgmtPortStr := os.Getenv("MANAGEMENT_SERVER_PORT")

	db, err := gorm.Open(sqlite.Open("dns.db"), &gorm.Config{})
	ctx := context.Background()

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&SecondLevelDomain{}, &A{}, &AAAA{}, &CNAME{}, &MX{}, &TXT{})

	sldsFromDB, _ := gorm.G[SecondLevelDomain](db).Find(ctx)

	for _, sld := range sldsFromDB {
		slds.PushBack(string(helpers.ConvertNameToBytes(sld.Name)))
	}

	generateHashTable(records, db, ctx)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/sld", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		var records []SecondLevelDomain

		db.Find(&records)

		enc.Encode(records)
	})

	r.Post("/sld", func(w http.ResponseWriter, r *http.Request) {
		record := SecondLevelDomain{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/sld", func(w http.ResponseWriter, r *http.Request) {
		record := SecondLevelDomain{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := SecondLevelDomain{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/sld", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&SecondLevelDomain{}, "name = ?", buf.String())
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Get("/a", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		var records []A

		db.Find(&records)

		enc.Encode(records)
	})

	r.Post("/a", func(w http.ResponseWriter, r *http.Request) {
		record := A{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/a", func(w http.ResponseWriter, r *http.Request) {
		record := A{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := A{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/a", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&A{}, "name = ?", buf.String())
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Get("/aaaa", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		var records []AAAA

		db.Find(&records)

		enc.Encode(records)
	})

	r.Post("/aaaa", func(w http.ResponseWriter, r *http.Request) {
		record := AAAA{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/aaaa", func(w http.ResponseWriter, r *http.Request) {
		record := AAAA{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := AAAA{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/aaaa", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&AAAA{}, "name = ?", buf.String())
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Get("/cname", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		var records []CNAME

		db.Find(&records)

		enc.Encode(records)
	})

	r.Post("/cname", func(w http.ResponseWriter, r *http.Request) {
		record := CNAME{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/cname", func(w http.ResponseWriter, r *http.Request) {
		record := CNAME{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := CNAME{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/cname", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&CNAME{}, "name = ?", buf.String())
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Get("/mx", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		var records []MX

		db.Find(&records)

		enc.Encode(records)
	})

	r.Post("/mx", func(w http.ResponseWriter, r *http.Request) {
		record := MX{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/mx", func(w http.ResponseWriter, r *http.Request) {
		record := MX{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := MX{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)
		incrementSerial()

		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/mx", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&MX{}, "name = ?", buf.String())
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Get("/txt", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		var records []TXT

		db.Find(&records)

		enc.Encode(records)
	})

	r.Post("/txt", func(w http.ResponseWriter, r *http.Request) {
		record := TXT{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/txt", func(w http.ResponseWriter, r *http.Request) {
		record := TXT{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := TXT{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)
		incrementSerial()

		db.First(&record, "name = ?", dbRecord.Name)
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/txt", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&TXT{}, "name = ?", buf.String())
		incrementSerial()
		generateHashTable(records, db, ctx)

		w.Write([]byte("OK"))
	})

	fmt.Printf("Management server running on http://%s:%s", mgmtIp, mgmtPortStr)
	http.ListenAndServe(fmt.Sprintf("%s:%s", mgmtIp, mgmtPortStr), r)

}
