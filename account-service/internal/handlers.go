package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
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

type RecordTransactionRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Kind        string `json:"kind"` // "credit" or "debit"
	Description string `json:"description"`
}

type TransactionResponse struct {
	ID          uint64 `json:"id"`
	AccountID   uint64 `json:"account_id"`
	AmountCents int64  `json:"amount_cents"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type TransactionsListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
}

type RecordActivityRequest struct {
	Action    string `json:"action"`
	IPAddress string `json:"ip_address"`
}

type ActivityResponse struct {
	ID        uint64 `json:"id"`
	AccountID uint64 `json:"account_id"`
	Action    string `json:"action"`
	IPAddress string `json:"ip_address"`
	CreatedAt string `json:"created_at"`
}

type ActivityListResponse struct {
	Activity []ActivityResponse `json:"activity"`
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

// handleAccountByID dispatches GET and DELETE for /accounts/{id}.
func (h *httpServer) handleAccountByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetAccount(w, r)
	case http.MethodDelete:
		h.handleDeleteAccount(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *httpServer) handleGetAccount(w http.ResponseWriter, r *http.Request) {
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

func (h *httpServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	accountID, err := parseAccountIDFromPath(r)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.svc.DeleteAccount(r.Context(), accountID)
	if err != nil {
		if err.Error() == "account not found" {
			writeJSONError(w, "not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, "failed to delete account", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *httpServer) handleRecordTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accountID, err := parseAccountIDFromPath(r)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req RecordTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Kind == "" {
		writeJSONError(w, "kind required (credit or debit)", http.StatusBadRequest)
		return
	}
	tx, err := h.svc.RecordTransaction(r.Context(), accountID, req.AmountCents, req.Kind, req.Description)
	if err != nil {
		if err.Error() == "kind must be credit or debit" {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONError(w, "failed to record transaction", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, TransactionResponse{
		ID:          tx.ID,
		AccountID:   tx.AccountID,
		AmountCents: tx.AmountCents,
		Kind:        tx.Kind,
		Description: tx.Description,
		CreatedAt:   tx.CreatedAt.Format(time.RFC3339),
	})
}

func (h *httpServer) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accountID, err := parseAccountIDFromPath(r)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	skip, take := parseSkipTake(r)
	txs, err := h.svc.ListTransactions(r.Context(), accountID, skip, take)
	if err != nil {
		writeJSONError(w, "failed to list transactions", http.StatusInternalServerError)
		return
	}
	list := make([]TransactionResponse, 0, len(txs))
	for _, t := range txs {
		list = append(list, TransactionResponse{
			ID:          t.ID,
			AccountID:   t.AccountID,
			AmountCents: t.AmountCents,
			Kind:        t.Kind,
			Description: t.Description,
			CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, TransactionsListResponse{Transactions: list})
}

func (h *httpServer) handleRecordActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accountID, err := parseAccountIDFromPath(r)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req RecordActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Action == "" {
		writeJSONError(w, "action required", http.StatusBadRequest)
		return
	}
	act, err := h.svc.RecordActivity(r.Context(), accountID, req.Action, req.IPAddress)
	if err != nil {
		writeJSONError(w, "failed to record activity", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, ActivityResponse{
		ID:        act.ID,
		AccountID: act.AccountID,
		Action:    act.Action,
		IPAddress: act.IPAddress,
		CreatedAt: act.CreatedAt.Format(time.RFC3339),
	})
}

func (h *httpServer) handleListActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accountID, err := parseAccountIDFromPath(r)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	skip, take := parseSkipTake(r)
	activities, err := h.svc.ListActivity(r.Context(), accountID, skip, take)
	if err != nil {
		writeJSONError(w, "failed to list activity", http.StatusInternalServerError)
		return
	}
	list := make([]ActivityResponse, 0, len(activities))
	for _, a := range activities {
		list = append(list, ActivityResponse{
			ID:        a.ID,
			AccountID: a.AccountID,
			Action:    a.Action,
			IPAddress: a.IPAddress,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, ActivityListResponse{Activity: list})
}

func parseAccountIDFromPath(r *http.Request) (uint64, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, errors.New("id required")
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func parseSkipTake(r *http.Request) (skip, take uint64) {
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
	return skip, take
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
