package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reqLine, err := parseRequestLine(string(data))
	if err != nil {
		return nil, err
	}
	req.RequestLine = reqLine
	return req, nil
}

// parseRequestLine parses the HTTP request line,
// checking for valid request line format and HTTP version.
// While ignoring the rest of the headers and body.
func parseRequestLine(line string) (RequestLine, error) {
	// Split HTTP request into parts
	lines := strings.Split(line, "\r\n")
	// Get Request-Line
	lines = strings.Split(lines[0], " ")
	if len(lines) != 3 {
		return RequestLine{}, errors.New("invalid request line")
	}
	// Extract HTTP version
	httpVersionParts := strings.Split(lines[2], "/")
	if len(httpVersionParts) != 2 {
		return RequestLine{}, errors.New("invalid HTTP version")
	}

	reqLine := RequestLine{
		HttpVersion:   httpVersionParts[1],
		RequestTarget: lines[1],
		Method:        lines[0],
	}

	// Verify METHOD is capitalized and alphabetical
	if !isAllCapsLetter(reqLine.Method) {
		return RequestLine{}, errors.New("invalid method")
	}

	// Verify HTTP version is 1.1
	if reqLine.HttpVersion != "1.1" {
		return RequestLine{}, errors.New("invalid HTTP version")
	}

	return reqLine, nil
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
