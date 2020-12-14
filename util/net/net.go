package net

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func HostPort(addr string, port interface{}) string {
	host := addr
	if strings.Count(addr, ":") > 0 {
		host = fmt.Sprintf("[%s]", addr)
	}
	if v, ok := port.(string); ok && v == "" {
		return host
	} else if v, ok := port.(int); ok && v == 0 && net.ParseIP(host) == nil {
		return host
	}

	return fmt.Sprintf("%s:%v", host, port)
}

// 寻找最小端口到最大端口范围内可以的端口
// Example: Listen("localhost:5000-6000", fn)
func Listen(addr string, fn func(string) (net.Listener, error)) (net.Listener, error) {
	if strings.Count(addr, ":") == 1 && strings.Count(addr, "-") == 0 {
		return fn(addr)
	}

	host, ports, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	prange := strings.Split(ports, "-")

	if len(prange) < 2 {
		// 单端口
		return fn(addr)
	}

	min, err := strconv.Atoi(prange[0])
	if err != nil {
		return nil, errors.New("unable to extract port range")
	}

	max, err := strconv.Atoi(prange[1])
	if err != nil {
		return nil, errors.New("unable to extract port range")
	}

	for port := min; port <= max; port++ {
		ln, err := fn(HostPort(host, port))
		if err == nil {
			return ln, nil
		}

		if port == max {
			return nil, err
		}
	}

	return nil, fmt.Errorf("unable to bind to %s", addr)
}
