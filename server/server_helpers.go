package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"willofdaedalus/superluminal/config"
)

func sendHeader(ctx context.Context, conn net.Conn, header string) (bool, error) {
	fmt.Println("sending header")
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(config.MaxConnectionTime))
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		modifiedHeader := fmt.Sprintf("%s.%d", header, time.Now().Unix())
		_, err := conn.Write([]byte(modifiedHeader))
		errChan <- err
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return false, fmt.Errorf("couldn't send header to client: %v", err)
		}
	case <-ctx.Done():
		return false, fmt.Errorf("context canceled before sending header to client")
	}

	return true, nil
}

func getIpAddr() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// skip loopback addresses
			if ip.IsLoopback() {
				continue
			}

			// handle both ipv4 and ipv6
			if ip.To4() != nil {
				// return first valid ipv4 address
				return ip.String(), nil
			} else if ip.To16() != nil {
				// return first valid ipv6 address
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IP address found")
}

func handleNewClient(conn net.Conn) {
	fmt.Printf("someone connected on %s\n", conn.RemoteAddr().String())

	_, err := conn.Write([]byte("hello and welcome to the session"))
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("sent welcome message to client")
}

func (s *Server) handleSignals() {
	s.sig = make(chan os.Signal, 1)
	signal.Notify(s.sig, s.signals...)

	go func() {
		for {
			switch <-s.sig {
			case syscall.SIGQUIT,
				syscall.SIGABRT,
				syscall.SIGTERM,
				syscall.SIGINT,
				syscall.SIGHUP:
				s.ShutdownServer()
				os.Exit(0)
			}
		}
	}()
}
