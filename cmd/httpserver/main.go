package main

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, myChunkHandler())
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func myHandler() func(w *response.Writer, req *request.Request) {
	return func(w *response.Writer, req *request.Request) {
		b := make([]byte, 0)
		h := headers.GetDefaultHeaders(0)

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			b = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
			w.WriteStatusLine(response.BadRequest)
		case "/myproblem":
			b = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

			w.WriteStatusLine(response.InternalServerError)
		default:
			b = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
			w.WriteStatusLine(response.OK)
		}

		h.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		h.Set("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(b)
	}
}

func myChunkHandler() func(w *response.Writer, req *request.Request) {
	return func(w *response.Writer, req *request.Request) {
		// Rerouting /httpbin/x to https://httpbin.org/x
		url := req.RequestLine.RequestTarget
		if !strings.HasPrefix(url, "/httpbin/") {
			return
		}
		url = "https://httpbin.org/" + strings.TrimPrefix(url, "/httpbin/")

		h := headers.GetDefaultHeaders(0)
		h.Set("Transfer-Encoding", "chunked")
		h.Remove("Content-Length")

		resp, err := http.Get(url)
		if err != nil {
			w.WriteStatusLine(response.InternalServerError)
			return
		}

		w.WriteStatusLine(response.OK)
		w.WriteHeaders(h)

		buff := make([]byte, 32) // replace it to 1024 after testing
		for {
			n, err := resp.Body.Read(buff)
			if n > 0 {
				fmt.Printf("Read %d bytes from response body\n", n)
				if _, err := w.WriteChunkedBody(buff[:n]); err != nil {
					return
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return
			}
		}
		w.WriteChunkedBodyDone()
	}
}
