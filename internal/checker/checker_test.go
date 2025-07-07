package checker_test

import (
	"testing"
	"time"

	"github.com/dedov-mm/netchecknova/internal/checker"
)

func TestCheckHostAndPort_TCP_Success(t *testing.T) {
	// Пример с работающим хостом и портом
	address := "http://google.com"
	host := "google.com"
	port := 80
	opts := checker.CheckOptions{
		PortTimeout:   2 * time.Second,
		ProxyAddress:  "",
	}

	result, err := checker.CheckHostAndPort(address, host, port, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.PortSuccess {
		t.Errorf("expected success, got failure: %v", result.PortError)
	}
	if result.Message == "" {
		t.Errorf("expected message, got empty")
	}
}

func TestCheckHostAndPort_TCP_Failure(t *testing.T) {
	address := "http://nonexistent.example.com"
	host := "nonexistent.example.com"
	port := 80
	opts := checker.CheckOptions{
		PortTimeout:   2 * time.Second,
		ProxyAddress:  "",
	}

	result, err := checker.CheckHostAndPort(address, host, port, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PortSuccess {
		t.Errorf("expected failure, got success")
	}
	if result.PortError == "" {
		t.Errorf("expected error message, got none")
	}
}
