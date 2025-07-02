package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dedov-mm/netchecknova/internal/checker"
	"net/url"
	"net"
)

// CheckRequest представляет тело запроса на проверку доступности хоста и порта.
// Если UseProxy == true, то для подключения используется SOCKS5 прокси по адресу Proxy.
type CheckRequest struct {
	Address  string `json:"address"`   // Адрес для проверки, включая порт, например "8.8.8.8:53" или "https://google.com:443"
	UseProxy bool   `json:"use_proxy"` // Флаг использования прокси
	Proxy    string `json:"proxy"`     // Адрес SOCKS5 прокси, например "127.0.0.1:1080"
}

func main() {
	e := echo.New()

	// Middleware для логирования запросов и обработки паник
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Раздача статических файлов фронтенда из папки "web"
	e.Static("/", "web")

	// POST /check — обработчик проверки доступности хоста и порта
	e.POST("/check", handleCheck)

	// Запуск HTTP-сервера на порту 8080
	e.Logger.Fatal(e.Start(":8080"))
}

// handleCheck обрабатывает POST-запрос на /check.
// Ожидает JSON тело с адресом, флагом use_proxy и опциональным прокси-адресом.
// Возвращает JSON с результатом проверки.
func handleCheck(c echo.Context) error {
	req := new(CheckRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Парсим адрес и порт из строки Address
	host, port, err := parseAddressWithPort(req.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid address: " + err.Error()})
	}

	opts := checker.DefaultCheckOptions()

	// Если включён прокси, проверяем наличие адреса и передаём в опции
	if req.UseProxy {
		if strings.TrimSpace(req.Proxy) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "proxy address is required when use_proxy is true"})
		}
		opts.ProxyAddress = req.Proxy
	}

	// Выполняем проверку
	result, err := checker.CheckHostAndPort(host, port, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// parseAddressWithPort разбирает строку адреса с портом,
// поддерживает форматы с http:// или https:// и обычные "host:port".
// Возвращает хост и порт как отдельные значения.
func parseAddressWithPort(input string) (string, int, error) {
    input = strings.TrimSpace(input)

    if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
        u, err := url.Parse(input)
        if err != nil {
            return "", 0, err
        }
        host := u.Hostname()
        portStr := u.Port()

        if portStr == "" {
            // Автоматически подставляем порт в зависимости от схемы
            switch u.Scheme {
            case "https":
                return host, 443, nil
            case "http":
                return host, 80, nil
            default:
                return "", 0, echo.NewHTTPError(http.StatusBadRequest, "unsupported scheme")
            }
        }

        port, err := net.LookupPort("tcp", portStr)
        if err != nil {
            return "", 0, err
        }
        return host, port, nil
    }

    // Формат host:port без схемы
    host, portStr, err := net.SplitHostPort(input)
    if err != nil {
        return "", 0, err
    }
    port, err := net.LookupPort("tcp", portStr)
    if err != nil {
        return "", 0, err
    }
    return host, port, nil
}
