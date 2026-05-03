package server

import (
	"fmt"
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
	s.closed.Store(true)
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
	rsp := []byte("HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello World!\n")

	_, err := conn.Write(rsp)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
	return
}
