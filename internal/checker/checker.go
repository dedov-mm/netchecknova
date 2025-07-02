package checker

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/proxy"
)

// CheckResult содержит результат проверки хоста и порта.
type CheckResult struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	PortSuccess bool      `json:"port_success"`
	PortError   string    `json:"port_error"`
}

// CheckOptions задаёт параметры проверки TCP.
type CheckOptions struct {
	PortTimeout  time.Duration // таймаут проверки TCP-порта
	ProxyAddress string        // адрес SOCKS5 прокси, если нужно
}

// DefaultCheckOptions возвращает параметры по умолчанию.
func DefaultCheckOptions() CheckOptions {
	return CheckOptions{
		PortTimeout: 3 * time.Second,
	}
}

// CheckHostAndPort проверяет доступность TCP-порта напрямую или через прокси.
//
// host - адрес хоста
// port - номер порта
// opts - параметры проверки (таймаут и прокси)
//
// Возвращает структуру CheckResult и ошибку.
func CheckHostAndPort(host string, port int, opts CheckOptions) (*CheckResult, error) {
	var dialer proxy.Dialer
	var err error

	if opts.ProxyAddress != "" {
		// Создаём SOCKS5-диалер через прокси
		dialer, err = proxy.SOCKS5("tcp", opts.ProxyAddress, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("не удалось создать SOCKS5 диалер: %v", err)
		}
	} else {
		// Прямое соединение
		dialer = &net.Dialer{
			Timeout: opts.PortTimeout,
		}
	}

	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	portSuccess := false
	var portErr string
	if err == nil {
		portSuccess = true
		_ = conn.Close()
	} else {
		portErr = err.Error()
	}

	result := &CheckResult{
		Host:        host,
		Port:        port,
		PortSuccess: portSuccess,
		PortError:   portErr,
	}
	return result, nil
}
