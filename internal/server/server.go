package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener    net.Listener
	HandlerFunc Handler
	closed      atomic.Bool
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode   int
	ErrorMessage string
}

func (he HandlerError) WriteTo(w io.Writer) error {
	if err := response.WriteStatusLine(w, response.StatusCode(he.StatusCode)); err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(he.ErrorMessage))
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}
	if _, err := w.Write([]byte(he.ErrorMessage)); err != nil {
		return err
	}
	return nil
}

func Serve(port int, handlerFunc Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := &Server{
		Listener:    ln,
		HandlerFunc: handlerFunc,
	}
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerErr := &HandlerError{
			StatusCode:   400,
			ErrorMessage: err.Error(),
		}
		if err := handlerErr.WriteTo(conn); err != nil {
			log.Printf("Error writing bad request response: %v", err)
			return
		}
		return
	}

	buff := bytes.NewBuffer([]byte{})
	if handlerErr := s.HandlerFunc(buff, req); handlerErr != nil {
		if err := handlerErr.WriteTo(conn); err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}
	} else {
		headers := response.GetDefaultHeaders(buff.Len())
		if err := response.WriteStatusLine(conn, response.OK); err != nil {
			log.Printf("Error writing status line: %v", err)
			return
		}
		if err := response.WriteHeaders(conn, headers); err != nil {
			log.Printf("Error writing headers: %v", err)
			return
		}
		if _, err := conn.Write(buff.Bytes()); err != nil {
			log.Printf("Error writing response body: %v", err)
			return
		}
	}
}
