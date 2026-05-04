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

type WriterState int

const (
	WriteStatusLineState = iota
	WriteHeadersState
	WriteBodyState
)

type Writer struct {
	IOWriter io.Writer
	State    WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		IOWriter: w,
		State:    WriteStatusLineState,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != WriteStatusLineState {
		return errors.New("wrong order, writer is not ready to write status line")
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

	_, err := w.IOWriter.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, respPhrase)))
	if err != nil {
		return err
	}
	w.State = WriteHeadersState
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != WriteHeadersState {
		return errors.New("wrong order, writer is not ready to write headers")
	}
	for key, value := range headers {
		_, err := w.IOWriter.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.IOWriter.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.State = WriteBodyState
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.State != WriteBodyState {
		return 0, errors.New("wrong order, writer is not ready to write body")
	}
	_, err := w.IOWriter.Write(body)
	if err != nil {
		return 0, err
	}
	return len(body), nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.State != WriteBodyState {
		return 0, errors.New("wrong order, writer is not ready to write chucked body")
	}
	// size of the chunk in hex
	_, err := w.IOWriter.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		return 0, err
	}
	// data of the chunk in given size
	_, err = w.IOWriter.Write(p)
	if err != nil {
		return 0, err
	}
	// end of the chunk
	_, err = w.IOWriter.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.State != WriteBodyState {
		return 0, errors.New("wrong order, writer is not ready to write chucked body")
	}
	if _, err := w.IOWriter.Write([]byte("0\r\n\r\n")); err != nil {
		return 0, err
	}
	return 0, nil
}
