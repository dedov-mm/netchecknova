package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/dedov-mm/netchecknova/internal/checker"
)

// CheckRequest представляет тело запроса для проверки доступности хоста и порта.
type CheckRequest struct {
	Address  string `json:"address"`   // Адрес для проверки (может быть ip:port или URL с http/https)
	UseProxy bool   `json:"use_proxy"` // Использовать ли HTTP proxy для проверки
	Proxy    string `json:"proxy"`     // Адрес HTTP proxy, например "proxy.example.com:3128"
}

func main() {
	e := echo.New()
	e.Static("/", "web") 
	e.POST("/check", handleCheck)
	e.Logger.Fatal(e.Start(":8080"))
}

// handleCheck обрабатывает HTTP POST запрос на проверку доступности хоста и порта.
// Принимает JSON с параметрами CheckRequest.
// Если в адресе отсутствует схема (http:// или https://), добавляет "http://" для проверки через HTTP proxy.
// При включённом прокси проверяет через него, иначе — обычное TCP соединение.
// Возвращает JSON с результатом проверки или ошибкой.
func handleCheck(c echo.Context) error {
	req := new(CheckRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	address := strings.TrimSpace(req.Address)
	if address == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "address required"})
	}

	if !addressHasScheme(address) {
		address = "http://" + address
	}

	host, port, err := parseAddressWithPort(address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid address: " + err.Error()})
	}

	opts := checker.DefaultCheckOptions()
	if req.UseProxy {
		proxyAddr := strings.TrimSpace(req.Proxy)
		if proxyAddr == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "proxy address is required"})
		}
		opts.ProxyAddress = proxyAddr
	}

	result, err := checker.CheckHostAndPort(address, host, port, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"results": []checker.CheckResult{*result}})
}

// addressHasScheme проверяет, содержит ли адрес схему http:// или https://.
func addressHasScheme(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// parseAddressWithPort разбирает адрес с URL-схемой и возвращает хост и порт.
// Если порт не указан, возвращает стандартный порт для схемы http (80) или https (443).
// Возвращает ошибку, если адрес невалидный или порт не числовой.
func parseAddressWithPort(address string) (host string, port int, err error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", 0, err
	}

	host = u.Hostname()
	pStr := u.Port()
	if pStr == "" {
		switch u.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		default:
			port = 0
		}
	} else {
		port, err = strconv.Atoi(pStr)
		if err != nil {
			return "", 0, err
		}
	}

	return host, port, nil
}
