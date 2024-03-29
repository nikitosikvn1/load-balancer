package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/roman-mazur/design-practice-2-template/httptools"
	"github.com/roman-mazur/design-practice-2-template/signal"
)

var (
	port = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 1, "request timeout time in seconds")
	https = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

var (
	timeout = time.Duration(*timeoutSec) * time.Second
	serversPool = []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}
	healthyServersMutex sync.Mutex
	healthyServers []string
)

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var client HttpClient = http.DefaultClient

func health(dst string, client HttpClient) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(dst string, rw http.ResponseWriter, r *http.Request, client HttpClient) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst

	resp, err := client.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", dst)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

func main() {
	flag.Parse()

	// TODO: Використовуйте дані про стан сервреа, щоб підтримувати список тих серверів, яким можна відправляти ззапит.
	for _, server := range serversPool {
		server := server

		checkServerHealth(server)
		go func() {
			for range time.Tick(10 * time.Second) {
				checkServerHealth(server)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// TODO: Рееалізуйте свій алгоритм балансувальника.
		healthyServersMutex.Lock()
		defer healthyServersMutex.Unlock()

		// Якщо немає доступних здорових серверів, повертаємо статус "Service Unavailable"
		if len(healthyServers) == 0 {
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		pathHash := hash(r.URL.Path)
		serverIndex := int(pathHash) % len(healthyServers)
		forward(healthyServers[serverIndex], rw, r, client)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}

// Function to check server availability
func checkServerHealth(server string) {
	isHealthy := health(server, client)
	log.Printf("\x1b[35m%s %t\x1b[0m", server, isHealthy)

	healthyServersMutex.Lock()
	defer healthyServersMutex.Unlock()

	index := -1
	for i, v := range healthyServers {
		if v == server {
			index = i
			break
		}
	}

	if isHealthy {
		if index == -1 {
			healthyServers = append(healthyServers, server)
		}
	} else {
		if index != -1 {
			healthyServers = append(healthyServers[:index], healthyServers[index+1:]...)
		}
	}
	fmt.Println(healthyServers)
}

// djb2 hash algorithm
func hash(s string) uint32 {
	var hash uint32
	for i := 0; i < len(s); i++ {
		hash += uint32(s[i])
		hash += (hash << 10)
		hash ^= (hash >> 6)
	}

	hash += (hash << 3)
	hash ^= (hash >> 11)
	hash += (hash << 15)

	return hash
}
