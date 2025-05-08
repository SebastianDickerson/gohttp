package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type Writer struct {
	io.Writer
	State WriterState
}

type WriterState int

const (
	WriterInit WriterState = iota
	WriterStatusLine
	WriterHeaders
	WriterBody
	WriterDone
)

type StatusCode int

const (
	StatusCodeOk          StatusCode = 200
	StatusCodeBadRequest  StatusCode = 400
	StatusCodeServerError StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	// map status code to string
	statusText := map[StatusCode]string{
		StatusCodeOk:          "OK",
		StatusCodeBadRequest:  "Bad Request",
		StatusCodeServerError: "Internal Server Error",
	}
	_, err := w.Write([]byte("HTTP/1.1 " + strconv.Itoa(int(statusCode)) + " " + statusText[statusCode] + "\r\n"))
	if err != nil {
		return err
	}
	return nil
}

// WriteHeaders writes the provided headers to the given writer.
func (w *Writer) WriteHeaders(h headers.Headers) error {

	for key, value := range h {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n")) // End of headers
	return err
}

func (w *Writer) WriteTrailer(h headers.Headers) error {
	for key, value := range h {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n")) // End of headers
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.Write(p)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkerBody(p []byte) (int, error) {
	n, err := w.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		return n, err
	}
	n, err = w.Write(p)
	if err != nil {
		return n, err
	}
	n, err = w.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}
	if n == 0 {
		w.WriteChunkedDone()
		return n, nil
	}
	return n, nil
}

func (w *Writer) WriteChunkedDone() (int, error) {
	n, err := w.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

// GetDefaultHeaders generates default HTTP headers, including Content-Length.
func GetDefaultHeaders(contentLength int) headers.Headers {
	return map[string]string{
		"Content-Type":   "text/plain",
		"Content-Length": fmt.Sprintf("%d", contentLength),
		"Connection":     "close",
	}
}
