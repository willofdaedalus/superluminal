package backend

import (
	"net"
	"time"

	"github.com/google/uuid"
)

func createClient(name string, conn net.Conn, isOwner bool) *sessionClient {
	return &sessionClient{
		name:    name,
		conn:    conn,
		uuid:    uuid.NewString(),
		joined:  time.Now(),
		isOwner: isOwner,
	}
}
