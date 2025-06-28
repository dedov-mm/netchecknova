package checker

import (
	"fmt"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type CheckResult struct {
	Host        string
	Port        int
	PingSuccess bool
	PingSummary string
	PortSuccess bool
	PortError   string
}

func CheckHostAndPort(host string, port int) (*CheckResult, error) {
	// Каналы для получения результата ping
	pingResultChan := make(chan *probing.Statistics, 1)
	pingErrorChan := make(chan error, 1)

	go func() {
		pinger, err := probing.NewPinger(host)
		if err != nil {
			pingErrorChan <- err
			return
		}
		pinger.Count = 3
		pinger.Timeout = 3 * time.Second
		pinger.SetPrivileged(true)
		err = pinger.Run()
		if err != nil {
			pingErrorChan <- err
			return
		}
		stats := pinger.Statistics()
		pingResultChan <- stats
	}()

	// Проверяем порт
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	portSuccess := false
	var portErr string
	if err == nil {
		portSuccess = true
		_ = conn.Close()
	} else {
		portErr = err.Error()
	}

	// Получаем результат ping
	var pingSuccess bool
	var pingSummary string
	select {
	case stats := <-pingResultChan:
		pingSuccess = stats.PacketLoss < 100
		pingSummary = fmt.Sprintf("Sent=%d Loss=%.2f%% AvgRTT=%v",
			stats.PacketsSent,
			stats.PacketLoss,
			stats.AvgRtt)
	case err := <-pingErrorChan:
		pingSuccess = false
		pingSummary = fmt.Sprintf("Ping error: %v", err)
	case <-time.After(5 * time.Second):
		pingSuccess = false
		pingSummary = "Ping timeout"
	}

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
