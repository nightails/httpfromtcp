package main

import (
	"crypto/sha256"
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
		// Rerouting /httpbin/html to https://httpbin.org/html
		url := req.RequestLine.RequestTarget
		if !strings.HasSuffix(url, "/httpbin/html") {
			return
		}
		url = "https://httpbin.org/html"

		h := headers.GetDefaultHeaders(0)
		h.Set("Transfer-Encoding", "chunked")
		h.Add("Trailer", "X-Content-SHA256")
		h.Add("Trailer", "X-Content-Length")
		h.Remove("Content-Length")

		resp, err := http.Get(url)
		if err != nil {
			w.WriteStatusLine(response.InternalServerError)
			return
		}
		defer resp.Body.Close()

		if err := w.WriteStatusLine(response.OK); err != nil {
			return
		}
		if err := w.WriteHeaders(h); err != nil {
			return
		}

		var body []byte
		buff := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buff)
			if n > 0 {
				chunk := buff[:n]
				body = append(body, chunk...)
				fmt.Printf("Read %d bytes from response body\n", n)
				if _, err := w.WriteChunkedBody(chunk); err != nil {
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

		hash := sha256.Sum256(body)
		bodyLen := len(body)

		th := headers.Headers{}
		th.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
		th.Set("X-Content-Length", fmt.Sprintf("%d", bodyLen))
		if err := w.WriteTrailers(th); err != nil {
			return
		}
	}
}
