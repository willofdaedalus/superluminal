package backend

import (
	"net"
	"strings"
	"testing"
)

func TestGetAllClients(t *testing.T) {
	in, _ := net.Pipe()

	mockClients := map[string]*sessionClient{
		"a": createClient("james", in, false),
		"b": createClient("obi wan", in, false),
		"c": createClient("peter parker", in, false),
		"d": createClient("captain america", in, false),
		"e": createClient("bruce wayne", in, false),
	}

	sess := session{
		clients:  mockClients,
		maxConns: uint8(len(mockClients)),
	}

	allClients := sess.GetAllClients()
	t.Log(strings.Join(allClients, "\n"))
}
