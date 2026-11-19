package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheck_StatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		status int
		wantUp bool
	}{
		{"200 is up", http.StatusOK, true},
		{"204 is up", http.StatusNoContent, true},
		{"404 is down", http.StatusNotFound, false},
		{"500 is down", http.StatusInternalServerError, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			res := Check(context.Background(), srv.Client(), srv.URL)

			if res.Err != nil {
				t.Fatalf("unexpected error: %v", res.Err)
			}
			if res.Up != tc.wantUp {
				t.Errorf("Up = %v, want %v", res.Up, tc.wantUp)
			}
			if res.StatusCode != tc.status {
				t.Errorf("StatusCode = %d, want %d", res.StatusCode, tc.status)
			}
		})
	}
}

func TestCheck_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res := Check(ctx, srv.Client(), srv.URL)

	if res.Up {
		t.Error("expected target to be down on timeout")
	}
	if res.Err == nil {
		t.Error("expected an error on timeout")
	}
}

func TestCheck_InvalidURL(t *testing.T) {
	res := Check(context.Background(), http.DefaultClient, "://bad-url")

	if res.Up || res.Err == nil {
		t.Errorf("expected failure for invalid url, got %+v", res)
	}
}
