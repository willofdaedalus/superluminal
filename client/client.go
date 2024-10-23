package client

import "net"

type Client struct {
	Name     string
	PassUsed string
}

func CreateClient(name string) *Client {
	return &Client{
		Name: name,
	}
}

func (c Client) ConnectToServer(host, port string) error {
	_, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	return nil
}
