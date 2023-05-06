package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var https = flag.Bool("https", false, "whether backends support HTTPs")

var serversPool = []string{
	"localhost:8080",
	"localhost:8081",
	"localhost:8082",
}

type report map[string][]string

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func main()  {
	flag.Parse()

	client := new(http.Client)
	client.Timeout = 10 * time.Second

	res := make([]report, len(serversPool))
	for i, s := range serversPool {
		resp, err := client.Get(fmt.Sprintf("%s://%s/report", scheme(), s))
		if err == nil {
			var data report
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				//log.Printf("error parsing froom %s: %s", s, err)
			} else {
				for k, v := range data {
					l := len(v)
					if l > 5 {
						l = 5
					}
					data[k] = v[len(v)-l:]
				}
				res[i] = data
			}
		} else {
			log.Printf("error %s %s", s, err)
		}

		log.Println("=========================")
		log.Println("SERVER", i, serversPool[i])
		log.Println("=========================")
		data, _ := json.MarshalIndent(res[i], "", "  ")
		log.Println(string(data))
	}
}
