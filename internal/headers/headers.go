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

	if key == "" {
		return 0, false, errors.New("invalid header: empty key")
	}
	if strings.ContainsAny(key, " \t") {
		return 0, false, errors.New("invalid header: whitespace in key")
	}

	h[key] = value

	return bytesConsumed, false, nil
}
