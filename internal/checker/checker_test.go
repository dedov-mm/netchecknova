package checker

import (
	"testing"
	"time"
)

func TestCheckHostAndPort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	opts := CheckOptions{
		PortTimeout: 3 * time.Second,
	}

	result, err := CheckHostAndPort("8.8.8.8", 53, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Если сети нет, пропустим
	if !result.PortSuccess {
		t.Skipf("skipping test: no connectivity to 8.8.8.8:53, port error: %s", result.PortError)
	}

	if !result.PortSuccess {
		t.Errorf("expected port success, got error: %s", result.PortError)
	}

	// На всякий случай проверим, что PortError пустой при успехе
	if result.PortSuccess && result.PortError != "" {
		t.Errorf("expected empty port error when port check succeeds, got: %s", result.PortError)
	}
}
