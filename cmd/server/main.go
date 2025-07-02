package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dedov-mm/netchecknova/internal/checker"
)

func main() {
	e := echo.New()

	// Middleware для логирования и восстановления после паники
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Отдаем фронт
	e.Static("/", "web")

	// Эндпоинт /check?host=xxx&port=yyy
	e.GET("/check", func(c echo.Context) error {
		host := c.QueryParam("host")
		portStr := c.QueryParam("port")

		if host == "" || portStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "host and port query parameters are required",
			})
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "port must be a valid integer",
			})
		}

		// Используем опции по умолчанию
		opts := checker.DefaultCheckOptions()

		result, err := checker.CheckHostAndPort(host, port, opts)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, result)
	})

	// Запускаем сервер на порту 8080
	e.Logger.Fatal(e.Start(":8080"))
}
