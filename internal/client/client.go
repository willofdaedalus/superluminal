package client

import "time"

type Client struct {
	Name       string    // name user sends to application
	Alive      bool      // client still connected
	Master     bool      // owner of the current
	TimeJoined time.Time // timestamp of when client connected
}

func CreateClient(name string, owner bool) *Client {
	return &Client{
		Name:       name,
		TimeJoined: time.Now(),
		Alive:      true,
		Master:     owner,
	}
}
