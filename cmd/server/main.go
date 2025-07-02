package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dedov-mm/netchecknova/internal/checker"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: netchecknova <host> <port>")
		os.Exit(1)
	}

	host := os.Args[1]

	var port int
	_, err := fmt.Sscanf(os.Args[2], "%d", &port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	// Используем параметры по умолчанию
	opts := checker.DefaultCheckOptions()

	// Запускаем проверку
	result, err := checker.CheckHostAndPort(host, port, opts)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Выводим результат
	fmt.Println("=== NetCheck Result ===")
	fmt.Printf("Host: %s\n", result.Host)
	fmt.Printf("Port: %d\n", result.Port)
	fmt.Printf("Ping success: %v\n", result.PingSuccess)
	fmt.Printf("Ping summary: %s\n", result.PingSummary)
	fmt.Printf("Port success: %v\n", result.PortSuccess)
	if !result.PortSuccess {
		fmt.Printf("Port error: %s\n", result.PortError)
	}
}
