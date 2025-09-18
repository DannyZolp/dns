package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"

	"dannyzolp.com/m/v2/helpers"
)

func generateHashTable(hashMap map[string][]byte, db *gorm.DB, ctx context.Context) {
	a, _ := gorm.G[A](db).Find(ctx)
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

	// create all A records
	for _, r := range a {
		helpers.CreateARecord(hashMap, r.Name, r.IP, r.TTL)
	}
}

func management(port int, ip string, hashMap map[string][]byte, wg *sync.WaitGroup) {
	defer wg.Done()

	db, err := gorm.Open(sqlite.Open("dns.db"), &gorm.Config{})
	ctx := context.Background()

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&A{}, &CNAME{}, &MX{}, &TXT{})

	generateHashTable(hashMap, db, ctx)

	r := chi.NewRouter()

	r.Post("/a", func(w http.ResponseWriter, r *http.Request) {
		record := A{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)

		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/a", func(w http.ResponseWriter, r *http.Request) {
		record := A{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := A{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)

		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Delete("/a", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&A{}, "name = ?", buf.String())
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Post("/cname", func(w http.ResponseWriter, r *http.Request) {
		record := CNAME{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/cname", func(w http.ResponseWriter, r *http.Request) {
		record := CNAME{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := CNAME{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)

		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})
	r.Delete("/cname", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&CNAME{}, "name = ?", buf.String())
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Post("/mx", func(w http.ResponseWriter, r *http.Request) {
		record := MX{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/mx", func(w http.ResponseWriter, r *http.Request) {
		record := MX{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := MX{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)

		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})
	r.Delete("/mx", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&MX{}, "name = ?", buf.String())
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Post("/txt", func(w http.ResponseWriter, r *http.Request) {
		record := TXT{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		db.Create(&record)
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	r.Patch("/txt", func(w http.ResponseWriter, r *http.Request) {
		record := TXT{}

		dec := json.NewDecoder(r.Body)
		dec.Decode(&record)

		dbRecord := TXT{}
		db.First(&dbRecord, "name = ?", record.Name)

		db.Model(&dbRecord).Updates(record)

		db.First(&record, "name = ?", dbRecord.Name)
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})
	r.Delete("/txt", func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)

		db.Delete(&TXT{}, "name = ?", buf.String())
		generateHashTable(hashMap, db, ctx)

		w.Write([]byte("OK"))
	})

	fmt.Printf("Management server running on http://%s:%d", ip, port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), r)

}
