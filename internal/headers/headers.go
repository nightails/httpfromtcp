package headers

import (
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	end := strings.Index(string(data), "\r\n")
	if end == -1 {
		// incomplete header, need more data
		return 0, false, err
	}
	if end == 0 {
		return len("\r\n"), true, nil
	}

	headerText := string(data[:end])
	bytesConsumed := end + len("\r\n")

	colonIndex := strings.Index(headerText, ":")
	if colonIndex == -1 {
		return 0, false, errors.New("invalid header: missing colon")
	}

	key := headerText[:colonIndex]
	value := strings.TrimSpace(headerText[colonIndex+1:])

	if !isValidFieldName(key) {
		return 0, false, errors.New("invalid header: invalid field name")
	}

	h[strings.ToLower(key)] = value

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
