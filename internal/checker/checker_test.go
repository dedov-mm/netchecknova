package checker

import (
	"testing"
	"time"
)

func TestCheckHostAndPort(t *testing.T) {
	opts := CheckOptions{
		PingCount:   2,
		PingTimeout: 2 * time.Second,
		PortTimeout: 2 * time.Second,
	}
	result, err := CheckHostAndPort("8.8.8.8", 53, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.PingSuccess {
		t.Errorf("expected ping success, got: %s", result.PingSummary)
	}
	if !result.PortSuccess {
		t.Errorf("expected port success, got error: %s", result.PortError)
	}
}
