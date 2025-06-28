package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dedov-mm/netchecknova/internal/checker"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: server <host> <port>")
		os.Exit(1)
	}

	host := os.Args[1]
	port := 0
	fmt.Sscanf(os.Args[2], "%d", &port)

	result, err := checker.CheckHostAndPort(host, port)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Ping success:", result.PingSuccess)
	fmt.Println("Ping summary:", result.PingSummary)
	fmt.Println("Port success:", result.PortSuccess)
	if !result.PortSuccess {
		fmt.Println("Port error:", result.PortError)
	}
}
