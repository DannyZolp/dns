package management

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func DynDnsService() {
	r := chi.NewRouter()
	db, err := gorm.Open(sqlite.Open("dns.db"), &gorm.Config{})
	ctx := context.Background()

	if err != nil {
		panic(err)
	}

	r.Get("/nic/update", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != os.Getenv("DYNDNS_USERNAME") || password != os.Getenv("DYNDNS_PASSWORD") {
			w.Write([]byte("badauth\n"))
			return
		}

		response := ""

		query := r.URL.Query()

		hostnames := strings.Split(query.Get("hostname"), ",")
		ip := query.Get("myip")

		for _, hostname := range hostnames {
			record, err := gorm.G[A](db).Where("name = ?", hostname).First(ctx)

			if err != nil {
				response += "nohost\n"
				continue
			}

			ra, err := gorm.G[A](db).Where("id = ?", record.ID).Update(ctx, "ip", ip)

			if ra != 1 || err != nil {
				response += "911\n"
				fmt.Printf("rowsAffected = %d, err = %v", ra, err)
				continue
			}

			response += "good\n"
		}

		w.Write([]byte(response))
	})
	http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("DNS_SERVER_IP"), os.Getenv("DYNDNS_PORT")), r)
}
