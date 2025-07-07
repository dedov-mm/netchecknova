package proxyhelper

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

// DialContextFunc - тип функции для DialContext
type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

// CreateAutoDialer возвращает функцию DialContext, которая:
// - определяет тип прокси по адресу,
// - возвращает соответствующий диалер.
//
// Если proxyAddr пустой, возвращает стандартный net.Dialer.
func CreateAutoDialer(proxyAddr string, timeout time.Duration) (DialContextFunc, error) {
	if proxyAddr == "" {
		// Без прокси
		return (&net.Dialer{
			Timeout: timeout,
		}).DialContext, nil
	}

	// Пробуем понять, HTTP CONNECT ли это
	isHTTP, err := isHTTPConnectProxy(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("не удалось определить тип прокси: %v", err)
	}
	if isHTTP {
		return httpConnectDialer(proxyAddr), nil
	}

	// Иначе считаем SOCKS5
	socksDialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать SOCKS5 диалер: %v", err)
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return socksDialer.Dial(network, address)
	}, nil
}

// isHTTPConnectProxy пробует сделать CONNECT и возвращает true, если это HTTP CONNECT прокси
func isHTTPConnectProxy(proxyAddr string) (bool, error) {
	conn, err := net.DialTimeout("tcp", proxyAddr, 3*time.Second)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	req := "CONNECT www.google.com:443 HTTP/1.1\r\nHost: www.google.com:443\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		return false, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == 200, nil
}

// httpConnectDialer возвращает функцию DialContext для HTTP CONNECT
func httpConnectDialer(proxyAddr string) DialContextFunc {
	return func(ctx context.Context, network, targetAddr string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, err
		}
		req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetAddr, targetAddr)
		if _, err := conn.Write([]byte(req)); err != nil {
			conn.Close()
			return nil, err
		}
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if resp.StatusCode != 200 {
			conn.Close()
			return nil, fmt.Errorf("proxy CONNECT failed: %s", resp.Status)
		}
		return conn, nil
	}
}
