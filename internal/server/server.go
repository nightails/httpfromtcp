package server

import "net"

type Server struct {
}

func Serve(port int) (*Server, error) {
	return &Server{}, nil
}

func (s *Server) Close() error {
	return nil
}

func (s *Server) listen() {
}

func (s *Server) handle(conn net.Conn) {
}
