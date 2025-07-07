package checker

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CheckOptions содержит параметры для проверки.
type CheckOptions struct {
	ProxyAddress string        // Адрес HTTP proxy (например, "proxy.example.com:3128"). Если пустой, проверка без прокси.
	PortTimeout  time.Duration // Таймаут для проверки TCP-порта и HTTP-запросов.
}

// DefaultCheckOptions возвращает параметры проверки по умолчанию.
func DefaultCheckOptions() CheckOptions {
	return CheckOptions{
		ProxyAddress: "",
		PortTimeout:  3 * time.Second,
	}
}

// CheckResult содержит результат проверки доступности хоста и порта.
type CheckResult struct {
	Host        string `json:"host"`                   // Хост (IP или домен)
	Port        int    `json:"port"`                   // Порт
	PortSuccess bool   `json:"port_success"`           // Флаг успешного TCP подключения
	PortError   string `json:"port_error,omitempty"`   // Ошибка TCP подключения, если есть

	HTTPChecked bool   `json:"http_checked"`           // Был ли выполнен HTTP-запрос
	HTTPStatus  int    `json:"http_status,omitempty"`  // HTTP статус ответа
	HTTPError   string `json:"http_error,omitempty"`   // Ошибка HTTP запроса, если есть

	Message string `json:"message,omitempty"`       // Текстовое сообщение с результатом проверки
}

// checkTCP проверяет доступность TCP-порта на заданном хосте с таймаутом.
// Возвращает true, если соединение успешно установлено, иначе false с ошибкой.
func checkTCP(host string, port int, timeout time.Duration) (bool, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false, err
	}
	conn.Close()
	return true, nil
}

// CheckHTTPViaProxy выполняет HTTP GET запрос к targetURL через HTTP proxy proxyAddr.
// Возвращает true и HTTP статус если запрос успешен (код 2xx или 3xx), иначе false и описание ошибки.
func CheckHTTPViaProxy(proxyAddr string, targetURL string, timeout time.Duration) (bool, int, string) {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   proxyAddr,
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return false, 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, resp.StatusCode, ""
	}

	return false, resp.StatusCode, fmt.Sprintf("HTTP status %d", resp.StatusCode)
}

// CheckHostAndPort проверяет доступность хоста и порта.
// Если указан ProxyAddress и адрес содержит схему (http/https), проверяет HTTP запросом через proxy.
// Иначе выполняет TCP проверку напрямую.
// Возвращает результат проверки или ошибку.
func CheckHostAndPort(address string, host string, port int, opts CheckOptions) (*CheckResult, error) {
	result := &CheckResult{
		Host: host,
		Port: port,
	}

	if opts.ProxyAddress != "" && hasScheme(address) {
		ok, status, httpErr := CheckHTTPViaProxy(opts.ProxyAddress, address, opts.PortTimeout)
		result.HTTPChecked = true
		result.HTTPStatus = status
		result.HTTPError = httpErr

		// TLS-сертификат не распознан, но соединение возможно
		if !ok && isTLSError(httpErr) {
			result.PortSuccess = true
			result.Message = "⚠️ Доступен, но TLS-сертификат не распознан"
			return result, nil
		}

		if ok {
			result.PortSuccess = true
			result.Message = buildHTTPStatusMessage(status)
		} else {
			result.PortSuccess = false
			result.Message = "❌ Недоступен: " + httpErr
			result.PortError = fmt.Sprintf("Ошибка HTTP-запроса через прокси %s: %s", opts.ProxyAddress, httpErr)
		}
		return result, nil
	}

	// TCP проверка
	ok, err := checkTCP(host, port, opts.PortTimeout)
	if ok {
		result.PortSuccess = true
		result.Message = "✅ Доступен"
	} else {
		result.PortSuccess = false
		if err != nil {
			result.PortError = err.Error()
			result.Message = "❌ Недоступен: " + err.Error()
		} else {
			result.PortError = "неизвестная ошибка TCP"
			result.Message = "❌ Недоступен: неизвестная ошибка TCP"
		}
	}

	return result, nil
}

// hasScheme проверяет, содержит ли строка схему "http://" или "https://".
func hasScheme(s string) bool {
	return len(s) >= 7 && (s[:7] == "http://" || (len(s) >= 8 && s[:8] == "https://"))
}

func buildHTTPStatusMessage(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "✅ Доступен"
	case status >= 300 && status < 400:
		return "⚠️ Доступен с перенаправлением"
	case status >= 400 && status < 500:
		return "⚠️ Доступен, но отказано в доступе"
	case status >= 500:
		return "⚠️ Доступен, но серверная ошибка"
	default:
		return fmt.Sprintf("⚠️ Доступен, но неизвестный статус HTTP %d", status)
	}
}

func isTLSError(errMsg string) bool {
	return strings.Contains(errMsg, "x509:") || strings.Contains(errMsg, "tls:")
}
