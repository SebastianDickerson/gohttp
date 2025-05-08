package request

import (
	"errors"
	"io"
	"log"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

const (
	CRLF       = "\r\n"
	bufferSize = 8
)

// Request represents an HTTP request with its request line.
type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	state       requestState
}

type requestState int

const (
	requestStateInit requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

// RequestLine represents the components of an HTTP request line.
type RequestLine struct {
	Method        string
	RequestTarget string
	HTTPVersion   string
}

// parse processes the input data and updates the Request state.
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		bytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			log.Println("Error parsing request:", err)
			return totalBytesParsed, err
		}
		if bytesParsed == 0 {
			break // Not enough data to parse
		}
		totalBytesParsed += bytesParsed
		if totalBytesParsed >= len(data) {
			break // All data has been parsed
		}
	}
	return totalBytesParsed, nil
}

// parseSingle processes a single step of the request parsing based on the current state.
func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInit:
		requestLine, bytesParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if bytesParsed == 0 {
			return 0, nil // Not enough data to parse
		}
		r.RequestLine = requestLine
		r.state = requestStateParsingHeaders
		return bytesParsed, nil

	case requestStateParsingHeaders:
		headersMap := headers.NewHeaders()
		bytesParsed, done, err := headersMap.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.Headers = headersMap
			if headersMap.Get("Content-Length") != "" {
				r.state = requestStateParsingBody
			} else {
				r.state = requestStateDone
			}
			return bytesParsed + len(CRLF), nil
		}
		return 0, nil // Not enough data to parse

	case requestStateParsingBody:
		contentLengthStr, err := strconv.Atoi(r.Headers["content-length"])
		if err != nil {
			return 0, errors.New("invalid Content-Length header")
		}

		remainingLength := contentLengthStr - len(r.Body)
		if remainingLength < 0 {
			return 0, errors.New("body length exceeds Content-Length header")
		}

		toAppend := data
		if len(data) > remainingLength {
			toAppend = data[:remainingLength]
		}

		r.Body = append(r.Body, toAppend...)
		if len(r.Body) > contentLengthStr {
			return 0, errors.New("body length exceeds Content-Length header")
		}

		if len(r.Body) == contentLengthStr {
			r.state = requestStateDone
		}

		return len(toAppend), nil

	case requestStateDone:
		return 0, errors.New("error: trying to read data in a done state")

	default:
		return 0, errors.New("error: unknown state")
	}
}

// parseRequestLine parses the request line into a RequestLine struct.
func parseRequestLine(requestLine string) (RequestLine, int, error) {
	lineEnd := strings.Index(requestLine, CRLF)
	if lineEnd == -1 {
		return RequestLine{}, 0, nil // Not enough data
	}

	line := requestLine[:lineEnd]
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return RequestLine{}, lineEnd + len(CRLF), errors.New("invalid request line format")
	}

	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")
	if httpVersion == parts[2] {
		return RequestLine{}, lineEnd + len(CRLF), errors.New("invalid HTTP version format")
	}

	return RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HTTPVersion:   httpVersion,
	}, lineEnd + len(CRLF), nil
}

// RequestFromReader reads and parses an HTTP request from an io.Reader.
func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{state: requestStateInit}

	for req.state != requestStateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += n

		bytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		if bytesParsed > 0 {
			copy(buf, buf[bytesParsed:readToIndex])
			readToIndex -= bytesParsed
		}
	}

	return req, nil
}
