package server

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"net/http"
	"sync/atomic"
)

type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

type Handler func(*response.Writer, *request.Request)

type ProxyHandler struct{}

type HandlerError struct {
	StatusCode int
	Message    string
}

// RequestHandler creates a handler that proxies requests to the specified target.
func (ph *ProxyHandler) RequestHandler(target string) Handler {
	return func(w *response.Writer, req *request.Request) {
		res, err := http.Get("http://httpbin.org" + target)
		if err != nil {
			log.Printf("Error making request to %s: %v", target, err)
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		defer res.Body.Close()

		headers := response.GetDefaultHeaders(0)
		delete(headers, "Content-Length")
		headers["Transfer-Encoding"] = "chunked"
		headers["Trailer"] = "X-Content-SHA256, X-Content-Length"

		if err := w.WriteStatusLine(response.StatusCode(res.StatusCode)); err != nil {
			log.Printf("Error writing status line: %v", err)
			return
		}

		if err := w.WriteHeaders(headers); err != nil {
			log.Printf("Error writing headers: %v", err)
			return
		}

		hash := sha256.New()
		totalBytes, err := streamResponseBody(res.Body, w, hash)
		if err != nil {
			log.Printf("Error streaming response body: %v", err)
			return
		}

		trailer := map[string]string{
			"X-Content-SHA256": fmt.Sprintf("%x", hash.Sum(nil)),
			"X-Content-Length": fmt.Sprintf("%d", totalBytes),
		}
		if err := w.WriteTrailer(trailer); err != nil {
			log.Printf("Error writing trailer: %v", err)
		}
	}
}

// Serve starts the server on the specified port and begins listening for connections.
func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		handler:  handler,
		listener: listener,
	}
	go server.listen()
	return server, nil
}

// Close shuts down the server and stops accepting new connections.
func (s *Server) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	return s.listener.Close()
}

// listen accepts incoming connections and handles them in separate goroutines.
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

// handle processes a single connection by sending a fixed HTTP response.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		writeError(&response.Writer{Writer: conn}, http.StatusBadRequest, err.Error())
		return
	}

	writer := &response.Writer{Writer: conn}
	s.handler(writer, req)
}

// writeError writes an error response to the client.
func writeError(w *response.Writer, statusCode int, message string) {
	hErr := &HandlerError{
		StatusCode: statusCode,
		Message:    message,
	}
	if err := hErr.Write(w); err != nil {
		log.Printf("Error writing error response: %v", err)
	}
}

// Write writes the error response to the client.
func (he *HandlerError) Write(w *response.Writer) error {
	if err := w.WriteStatusLine(response.StatusCode(he.StatusCode)); err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(he.Message))
	if err := w.WriteHeaders(headers); err != nil {
		return err
	}

	_, err := w.Write([]byte(he.Message))
	return err
}

// streamResponseBody streams the response body in chunks and calculates its SHA256 hash.
func streamResponseBody(body io.Reader, w *response.Writer, hash io.Writer) (int, error) {
	buf := make([]byte, 1024)
	totalBytes := 0

	for {
		n, err := body.Read(buf)
		if n > 0 {
			totalBytes += n
			if _, writeErr := w.WriteChunkerBody(buf[:n]); writeErr != nil {
				return totalBytes, writeErr
			}
			if _, hashErr := hash.Write(buf[:n]); hashErr != nil {
				return totalBytes, hashErr
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return totalBytes, err
		}
	}

	if _, err := w.WriteChunkedDone(); err != nil {
		return totalBytes, err
	}

	return totalBytes, nil
}
