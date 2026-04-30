package headers

import (
	"errors"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIndex := strings.Index(string(data), crlf)
	if crlfIndex == -1 {
		// incomplete header, need more data
		return 0, false, err
	}
	if crlfIndex == 0 {
		// The empty line. Headers are done.
		return len(crlf), true, nil
	}

	headerText := string(data[:crlfIndex])
	bytesConsumed := crlfIndex + len(crlf)

	colonIndex := strings.Index(headerText, ":")
	if colonIndex == -1 {
		return 0, false, errors.New("invalid header: missing colon")
	}

	key := headerText[:colonIndex]
	value := strings.TrimSpace(headerText[colonIndex+1:])

	if !isValidFieldName(key) {
		return 0, false, errors.New("invalid header: invalid field name")
	}

	key = strings.ToLower(key)
	if existingValue, exists := h[key]; exists {
		value = existingValue + ", " + value
	}

	h[key] = value

	return bytesConsumed, false, nil
}

func isValidFieldName(key string) bool {
	if key == "" {
		return false
	}

	for i := 0; i < len(key); i++ {
		if !isTokenChar(key[i]) {
			return false
		}
	}

	return true
}

func isTokenChar(c byte) bool {
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}

	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}
