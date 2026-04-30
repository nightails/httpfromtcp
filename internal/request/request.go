package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	parseState  parseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parseState int

const (
	initialized parseState = iota
	done
)

const bufferSize = 1024

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		parseState: initialized,
	}

	for req.parseState != done {
		if readToIndex == len(buf) {
			newSize := len(buf) * 2
			if newSize == 0 {
				newSize = bufferSize
			}
			buf2 := make([]byte, newSize)
			copy(buf2, buf)
			buf = buf2
		}

		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n

		if readToIndex > 0 {
			np, parseErr := req.parse(buf[:readToIndex])
			if parseErr != nil {
				return nil, parseErr
			}

			if np > 0 {
				readToIndex -= np
				buf2 := make([]byte, max(bufferSize, readToIndex))
				copy(buf2, buf[np:np+readToIndex])
				buf = buf2
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				req.parseState = done
				break
			}
			return nil, err
		}
	}

	return req, nil
}

// parseRequestLine parses the HTTP request line and returns the request line,
// the number of bytes it consumed, plus any error.
func parseRequestLine(line string) (RequestLine, int, error) {
	end := strings.Index(line, "\r\n")
	if end == -1 {
		// Need more data before we can parse the request line
		return RequestLine{}, 0, nil
	}

	requestLineText := line[:end]
	bytesConsumed := end + len("\r\n")

	parts := strings.Split(requestLineText, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("invalid request line")
	}

	httpVersionParts := strings.Split(parts[2], "/")
	if len(httpVersionParts) != 2 {
		return RequestLine{}, 0, errors.New("invalid HTTP version")
	}
	if httpVersionParts[0] != "HTTP" {
		return RequestLine{}, 0, errors.New("invalid HTTP version")
	}

	reqLine := RequestLine{
		HttpVersion:   httpVersionParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	if !isAllCapsLetter(reqLine.Method) {
		return RequestLine{}, 0, errors.New("invalid method")
	}
	if reqLine.HttpVersion != "1.1" {
		return RequestLine{}, 0, errors.New("invalid HTTP version")
	}

	return reqLine, bytesConsumed, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.parseState {
	case initialized:
		reqLine, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = reqLine
		r.parseState = done
		return n, nil
	case done:
		return 0, errors.New("request already parsed")
	default:
		return 0, errors.New("invalid parse state")
	}
}

// isAllCapsLetter returns true if the string is all capital letters
func isAllCapsLetter(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
