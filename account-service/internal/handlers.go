package internal

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Request/response DTOs for REST API (no password in responses).

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type AccountResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AccountsListResponse struct {
	Accounts []AccountResponse `json:"accounts"`
}

func (h *httpServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeJSONError(w, "name, email and password required", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if err.Error() == "account already exists" {
			writeJSONError(w, err.Error(), http.StatusConflict)
			return
		}
		writeJSONError(w, "registration failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, TokenResponse{Token: token})
}

func (h *httpServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "email and password required", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "account not found" {
			writeJSONError(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		writeJSONError(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, TokenResponse{Token: token})
}

func (h *httpServer) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		writeJSONError(w, "id required", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	account, err := h.svc.GetAccount(r.Context(), id)
	if err != nil {
		writeJSONError(w, "not found", http.StatusNotFound)
		return
	}
	if account == nil {
		writeJSONError(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, AccountResponse{
		ID:    account.ID,
		Name:  account.Name,
		Email: account.Email,
	})
}

func (h *httpServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var skip, take uint64
	if s := r.URL.Query().Get("skip"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			skip = v
		}
	}
	if t := r.URL.Query().Get("take"); t != "" {
		if v, err := strconv.ParseUint(t, 10, 64); err == nil {
			take = v
		}
	}
	if take == 0 {
		take = 100
	}

	accounts, err := h.svc.GetAccounts(r.Context(), skip, take)
	if err != nil {
		writeJSONError(w, "failed to list accounts", http.StatusInternalServerError)
		return
	}

	list := make([]AccountResponse, 0, len(accounts))
	for _, a := range accounts {
		list = append(list, AccountResponse{ID: a.ID, Name: a.Name, Email: a.Email})
	}
	writeJSON(w, http.StatusOK, AccountsListResponse{Accounts: list})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
