package checker

import (
	"fmt"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// CheckResult содержит результат проверки хоста и порта.
type CheckResult struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	PingSuccess bool   `json:"ping_success"`
	PingSummary string `json:"ping_summary"`
	PortSuccess bool   `json:"port_success"`
	PortError   string `json:"port_error"`
}

// CheckOptions задаёт параметры проверки ICMP и TCP.
type CheckOptions struct {
	PingCount   int           // сколько пакетов отправлять при пинге
	PingTimeout time.Duration // таймаут пинга
	PortTimeout time.Duration // таймаут проверки TCP-порта
}

// DefaultCheckOptions возвращает параметры по умолчанию.
func DefaultCheckOptions() CheckOptions {
	return CheckOptions{
		PingCount:   3,
		PingTimeout: 3 * time.Second,
		PortTimeout: 3 * time.Second,
	}
}

// CheckHostAndPort проверяет доступность хоста по ICMP (ping) и TCP (port).
//
// host - адрес хоста
// port - номер порта
// opts - параметры проверки
//
// Возвращает структуру CheckResult и ошибку (если произошла системная ошибка).
func CheckHostAndPort(host string, port int, opts CheckOptions) (*CheckResult, error) {
	// Каналы для получения результата ping
	pingResultChan := make(chan *probing.Statistics, 1)
	pingErrorChan := make(chan error, 1)

	// Запускаем пинг в отдельной горутине
	go func() {
		pinger, err := probing.NewPinger(host)
		if err != nil {
			pingErrorChan <- err
			return
		}
		pinger.Count = opts.PingCount
		pinger.Timeout = opts.PingTimeout
		pinger.SetPrivileged(false) // без прав root
		err = pinger.Run()
		if err != nil {
			pingErrorChan <- err
			return
		}
		stats := pinger.Statistics()
		pingResultChan <- stats
	}()

	// Проверка TCP-порта
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, opts.PortTimeout)
	portSuccess := false
	var portErr string
	if err == nil {
		portSuccess = true
		_ = conn.Close()
	} else {
		portErr = err.Error()
	}

	// Получение результата пинга
	var pingSuccess bool
	var pingSummary string
	select {
	case stats := <-pingResultChan:
		pingSuccess = stats.PacketLoss < 100
		pingSummary = fmt.Sprintf(
			"Sent=%d Loss=%.2f%% AvgRTT=%v",
			stats.PacketsSent,
			stats.PacketLoss,
			stats.AvgRtt,
		)
	case err := <-pingErrorChan:
		pingSuccess = false
		pingSummary = fmt.Sprintf("Ping error: %v", err)
	case <-time.After(opts.PingTimeout + 2*time.Second):
		pingSuccess = false
		pingSummary = "Ping timeout"
	}

	// Сбор финального результата
	result := &CheckResult{
		Host:        host,
		Port:        port,
		PingSuccess: pingSuccess,
		PingSummary: pingSummary,
		PortSuccess: portSuccess,
		PortError:   portErr,
	}
	return result, nil
}
