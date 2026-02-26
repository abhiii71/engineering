package internal

import (
	"fmt"
	"net/http"
)

type httpServer struct {
	svc  AccountService
	port int
}

// ListenREST starts the REST API server (blocks).
func ListenREST(svc AccountService, port int) error {
	h := &httpServer{svc: svc, port: port}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.handleRegister)
	mux.HandleFunc("POST /login", h.handleLogin)
	mux.HandleFunc("GET /accounts/{id}", h.handleGetAccount)
	mux.HandleFunc("GET /accounts", h.handleGetAccounts)
	mux.HandleFunc("POST /accounts/{id}/transactions", h.handleRecordTransaction)
	mux.HandleFunc("GET /accounts/{id}/transactions", h.handleListTransactions)
	mux.HandleFunc("POST /accounts/{id}/activity", h.handleRecordActivity)
	mux.HandleFunc("GET /accounts/{id}/activity", h.handleListActivity)

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}
