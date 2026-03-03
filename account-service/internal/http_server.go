package internal

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type httpServer struct {
	svc  AccountService
	port int
}

// responseWriter wraps http.ResponseWriter to capture status code and size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// logRequest logs each request: method, path, remote addr, status, duration, size.
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		dur := time.Since(start)
		log.Printf("%s %s %s %d %d %s", r.Method, r.URL.Path, r.RemoteAddr, wrapped.status, wrapped.size, dur)
	})
}

// ListenREST starts the REST API server (blocks).
func ListenREST(svc AccountService, port int) error {
	h := &httpServer{svc: svc, port: port}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.handleRegister)
	mux.HandleFunc("POST /login", h.handleLogin)
	mux.HandleFunc("GET /accounts", h.handleGetAccounts)
	mux.HandleFunc("/accounts/{id}", h.handleAccountByID)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Account service (REST API) listening on %s", addr)
	return http.ListenAndServe(addr, logRequest(mux))
}
