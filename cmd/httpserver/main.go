package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"maps"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()

	log.Printf("Server started on port %d", port)

	waitForShutdown()
	log.Println("Server gracefully stopped")
}

func handleRequest(w *response.Writer, r *request.Request) {
	path := r.RequestLine.RequestTarget

	if strings.HasPrefix(path, "/httpbin") {
		trimedPath := strings.TrimPrefix(path, "/httpbin")
		proxyHandler := server.ProxyHandler{}
		handler := proxyHandler.RequestHandler(trimedPath)
		handler(w, r)
		return
	}

	switch path {
	case "/yourproblem":
		respondWithHTML(w, response.StatusCodeBadRequest, "400 Bad Request", "Your request honestly kinda sucked.")
	case "/myproblem":
		respondWithHTML(w, response.StatusCodeServerError, "500 Internal Server Error", "Okay, you know what? This one is on me.")
	default:
		respondWithHTML(w, response.StatusCodeOk, "200 OK", "Your request was an absolute banger.", map[string]string{"LETSGO": "YES"})
	}
}

func respondWithHTML(w *response.Writer, statusCode response.StatusCode, title, message string, extraHeaders ...map[string]string) {
	body := []byte(`
<html>
	<head>
		<title>` + title + `</title>
	</head>
	<body>
		<h1>` + title + `</h1>
		<p>` + message + `</p>
	</body>
</html>
	`)

	w.WriteStatusLine(statusCode)

	headers := response.GetDefaultHeaders(len(body))
	headers["Content-Type"] = "text/html"

	if len(extraHeaders) > 0 {
		maps.Copy(headers, extraHeaders[0])
	}

	w.WriteHeaders(headers)
	w.WriteBody(body)
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
