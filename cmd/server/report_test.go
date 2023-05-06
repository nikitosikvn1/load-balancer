package main

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestReport_Process(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("lb-author", "test-author")
	req.Header.Set("lb-req-cnt", "1")

	r := make(Report)

	r.Process(req)
	if !reflect.DeepEqual(r["test-author"], []string{"1"}) {
		t.Errorf("Unexpected report state %s", r)
	}

	req.Header.Set("lb-req-cnt", "2")
	r.Process(req)
	if !reflect.DeepEqual(r["test-author"], []string{"1", "2"}) {
		t.Errorf("Unexpected report state %s", r)
	}

	req.Header.Set("lb-author", "test-len")
	for i := 0; i < 103; i++ {
		req.Header.Set("lb-req-cnt", "test-len")
		r.Process(req)
	}
	if len(r["test-len"]) != reportMaxLen {
		t.Errorf("Unexpectd error length: %d", len(r["test-len"]))
	}
}
