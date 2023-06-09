package integration

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	baseAddress      = "http://balancer:8090"
	numberOfRequests = 3
)

var serversPool = []string{
	"server1:8080",
	"server2:8080",
	"server3:8080",
}

var client = http.Client{
	Timeout: 1 * time.Second,
}

func TestBalancer(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}

	flag.Parse()

	var data = [3]string{"v1/capitals/berlin", "v1/planets/earth", "v1/data/qwerty"}

	serverResponses := make(map[string]int)

	for i := 0; i < numberOfRequests; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/%s", baseAddress, data[i]))
		if err != nil {
			t.Error(err)
			return
		}

		server := resp.Header.Get("lb-from")
		t.Logf("response from [%s]", server)
		if server == "" {
			t.Error("No lb-from header in response")
		} else if server != serversPool[i] {
			t.Errorf("Request was sent not to the expected server, sent to %s, expected %s", server, serversPool[i])
		}

		serverResponses[server]++
	}

	if len(serverResponses) != 3 {
		t.Errorf("Expected responses from %v servers, but got responses from %v servers", len(serversPool), len(serverResponses))
		return
	}

	t.Log("Balancer distributed requests among the following servers:")
}

func BenchmarkBalancer(b *testing.B) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		b.Skip("Integration test is not enabled")
	}

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			b.Error("Error in benchmark: ", err)
		}
		resp.Body.Close()
	}
}
