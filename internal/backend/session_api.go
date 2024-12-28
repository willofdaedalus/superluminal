package backend

import "fmt"

// name, ipaddr, timejoined
func (s *Session) GetAllClients() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	allClients := make([]string, len(s.clients))
	for _, v := range s.clients {
		if !v.isOwner {
			client := fmt.Sprintf("%s$$%s$$%s",
				v.name,
				v.conn.RemoteAddr().String(),
				v.joined.String(),
			)
			allClients = append(allClients, client)
		}
	}

	return allClients
}

func (s *Session) GetClientCount() string {
	return fmt.Sprintf("%d / %d", len(s.clients), s.maxConns)
}

func (s *Session) GetCurrentPass() string {
	return s.pass
}
