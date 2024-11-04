package backend

func (s *Server) GetMaxConns() int {
	return s.maxClients
}
