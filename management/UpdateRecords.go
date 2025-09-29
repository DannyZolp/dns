package management

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func databaseBroadcast() {
	l, err := net.Listen("tcp", net.JoinHostPort(os.Getenv("MANAGEMENT_SERVER_IP"), os.Getenv("MANAGEMENT_SERVER_PORT")))

	if err != nil {
		panic(err)
	}

	for {
		c, err := l.Accept()

		if err != nil {
			fmt.Println(err)
			c.Close()
			continue
		}

		db, err := os.ReadFile("dns.db")
		enc := base64.StdEncoding.EncodeToString(db) + "\n"

		if err != nil {
			fmt.Println(err)
			c.Close()
			continue
		}

		c.Write([]byte(enc))
		c.Close()
	}
}

func UpdateRecords(records map[string][]byte, wg *sync.WaitGroup) {
	defer wg.Wait()
	if os.Getenv("TYPE") == "root" {
		fmt.Println("Broadcasting database...")
		go databaseBroadcast()

		for {
			GenerateRecordMap(records)

			time.Sleep(time.Second * 60)
		}
	} else if os.Getenv("TYPE") == "child" {
		for {
			c, err := net.Dial("tcp", net.JoinHostPort(os.Getenv("MANAGEMENT_SERVER_IP"), os.Getenv("MANAGEMENT_SERVER_PORT")))
			if err != nil {
				fmt.Println(err)
				time.Sleep(time.Second * 60)
				continue
			}
			reader := bufio.NewReader(c)

			data, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				time.Sleep(time.Second * 60)
				continue
			}

			dec, _ := base64.StdEncoding.DecodeString(data)

			os.WriteFile("dns.db", dec, 0666)

			GenerateRecordMap(records)

			time.Sleep(time.Second * 300)
		}
	}
}
