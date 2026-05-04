package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

// Parse parses the given data and populates the Headers map with key-value pairs.
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// print the data with crlf encoding

	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := strings.ToLower(string(parts[0]))

	if key != strings.TrimSpace(key) {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}
	h.Add(key, string(value))
	return idx + 2, false, nil
}

// Add adds or updates a header with the specified key and value. If the key exists, values are concatenated with a comma.
func (h Headers) Add(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

// Set updates or assigns the specified key and value in the Headers map, converting the key to lowercase.
func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

// Get retrieves the value associated with the specified key in a case-insensitive manner from the Headers map.
func (h Headers) Get(key string) string {
	v, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return v
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

// validTokens checks if the data contains only valid tokens
// or characters that are allowed in a token
func validTokens(data []byte) bool {
	for _, c := range data {
		if !(c >= 'A' && c <= 'Z' ||
			c >= 'a' && c <= 'z' ||
			c >= '0' && c <= '9' ||
			c == '-') {
			return false
		}
	}
	return true
}
