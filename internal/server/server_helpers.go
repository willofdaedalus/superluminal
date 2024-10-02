package server

import (
	"fmt"
	"log"
	"net"
)

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

			// Skip loopback addresses
			if ip.IsLoopback() {
				continue
			}

			// Handle both IPv4 and IPv6
			if ip.To4() != nil {
				// Return first valid IPv4 address
				return ip.String(), nil
			} else if ip.To16() != nil {
				// Return first valid IPv6 address
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IP address found")
}

func handleNewClient(conn net.Conn) {
	defer conn.Close()

	clientData := make([]byte, 1024)
	_, err := conn.Read(clientData)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("got %s from the client\n", clientData)
}
