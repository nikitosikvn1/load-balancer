package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var target = flag.String("target", "http://localhost:8090", "request target")

func main() {
	flag.Parse()
	client := new(http.Client)
	client.Timeout = 10 * time.Second

	for range time.Tick(1 * time.Second) {
		rroute := GenerateRandomRoute()
		resp, err := client.Get(fmt.Sprintf("%s/%s", *target, rroute))
		if err == nil {
			log.Printf("response %d", resp.StatusCode)
		} else {
			log.Printf("error %s", err)
		}
	}
}

// Function to generate random URL routes (used for load balancer test)
func GenerateRandomRoute() string {
	rand.Seed(time.Now().UnixNano())

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	length := rand.Intn(8) + 3
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		result[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(result)
}
