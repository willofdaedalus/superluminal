package backend

import (
	"net"
	"time"

	"github.com/google/uuid"
)

func createClient(name, pass string, conn net.Conn, isOwner bool) *sessionClient {
	return &sessionClient{
		name:    name,
		pass:    pass,
		conn:    conn,
		uuid:    uuid.NewString(),
		joined:  time.Now(),
		isOwner: isOwner,
	}
}
