package response

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}
	if statusCode < 100 || statusCode >= 600 {
		return errors.New("invalid status code")
	}

	respPhrase := ""
	switch statusCode {
	case OK:
		respPhrase = "OK"
	case BadRequest:
		respPhrase = "Bad Request"
	case InternalServerError:
		respPhrase = "Internal Server Error"
	}

	respLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, respPhrase)
	if _, err := w.Write([]byte(respLine)); err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%d", contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}
	for key, value := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		if _, err := w.Write([]byte(headerLine)); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}
	return nil
}
