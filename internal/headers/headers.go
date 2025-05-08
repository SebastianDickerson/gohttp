package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

const CRLF = "\r\n"

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Convert data to string for easier manipulation
	dataStr := string(data)

	for !done {
		// Find the end of the header line
		endOfLine := strings.Index(dataStr, CRLF)
		if endOfLine == -1 {
			return 0, false, nil
		}

		headerLine := dataStr[:endOfLine]
		dataStr = dataStr[endOfLine+len(CRLF):]
		n += endOfLine + len(CRLF)

		if headerLine == "" {
			// Empty line indicates the end of headers
			done = true
			break
		}
		// Split the header line into key and value
		colonIndex := strings.Index(headerLine, ":")
		if colonIndex == -1 {
			return 0, false, fmt.Errorf("invalid header line: %s", headerLine)
		}
		if headerLine[colonIndex-1] == ' ' {
			return 0, false, fmt.Errorf("invalid header line: %s", headerLine)
		}

		key := strings.TrimSpace(strings.ToLower(headerLine[:colonIndex]))
		value := strings.TrimSpace(headerLine[colonIndex+1:])

		if !isValidHeaderFieldName(key) {
			return 0, false, fmt.Errorf("invalid header field name: %s", key)
		}

		if key == "" || value == "" {
			return 0, false, fmt.Errorf("invalid header line: %s", headerLine)
		}

		if _, exists := h[key]; exists {
			// If the header already exists, append the new value
			h[key] += ", " + value
		} else {
			// If the header does not exist, add it to the map
			h[key] = value
		}
	}

	return n - len(CRLF), true, nil
}

func (h Headers) Get(key string) string {
	if value, exists := h[strings.ToLower(key)]; exists {
		return value
	}
	return ""
}

func isValidHeaderFieldName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, c := range name {
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '!' || c == '#' || c == '$' || c == '%' || c == '&' ||
			c == '\'' || c == '*' || c == '+' || c == '-' || c == '.' ||
			c == '^' || c == '_' || c == '`' || c == '|' || c == '~':
		default:
			return false
		}
	}
	return true
}
