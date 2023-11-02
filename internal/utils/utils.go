// Package utils содержит различные функции общего назначения.
package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"flag"
	"net"
	"time"
)

// Wait предназначена для выполнения ожидания в течение d, или пока не будет прерван контекст.
func Wait(ctx context.Context, d time.Duration) error {
	if d == 0 {
		return ctx.Err()
	}

	ticker := time.NewTicker(d)

	select {
	case <-ticker.C:
		ticker.Stop()

		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// HashSHA256 производить вычисление SHA256 хеша.
func HashSHA256(src, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	_, err := h.Write(src)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// IsFlagPassed проверяет, был ли указан флаг запуска.
func IsFlagPassed(name string) bool {
	found := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

func LocalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if isPrivateIP(ip) {
				return ip, nil
			}
		}
	}

	return nil, errors.New("no IP")
}

func isPrivateIP(ip net.IP) bool {
	var privateIPBlocks []*net.IPNet
	for _, cidr := range []string{
		// don't check loopback ips
		//"127.0.0.0/8",    // IPv4 loopback
		//"::1/128",        // IPv6 loopback
		//"fe80::/10",      // IPv6 link-local
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}
