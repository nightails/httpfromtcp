package server

import (
	"fmt"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := &Server{Listener: ln}
	go srv.listen()
	return srv, nil
}

func (s *Server) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	if err := response.WriteStatusLine(conn, response.OK); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
}
