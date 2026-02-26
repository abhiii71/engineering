package internal

import (
	"context"
	"database/sql"

	model "github.com/abhiii71/engineering/account-service/models"
)

type AccountRepository interface {
	Close() error
	PutAccount(ctx context.Context, a model.Account) (*model.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*model.Account, error)
	GetAccountByID(ctx context.Context, id uint64) (*model.Account, error)
	ListAccounts(ctx context.Context, skip, take uint64) ([]model.Account, error)
	// Transactions
	PutTransaction(ctx context.Context, t model.Transaction) (*model.Transaction, error)
	ListTransactions(ctx context.Context, accountID uint64, skip, take uint64) ([]model.Transaction, error)
	// Activity log
	PutActivityLog(ctx context.Context, a model.ActivityLog) (*model.ActivityLog, error)
	ListActivityLog(ctx context.Context, accountID uint64, skip, take uint64) ([]model.ActivityLog, error)
}

type repo struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &repo{db: db}
}

func (r *repo) Close() error {
	return r.db.Close()
}

func (r *repo) PutAccount(ctx context.Context, a model.Account) (*model.Account, error) {
	query := `Insert into accounts (name, email, password) VALUES($1, $2, $3) RETURNING id`

	err := r.db.QueryRowContext(ctx, query, a.Name, a.Email, a.Password).Scan(&a.ID)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repo) GetAccountByEmail(ctx context.Context, email string) (*model.Account, error) {
	var account model.Account
	query := `Select id, name, email, password FROM accounts where email=$1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&account.ID, &account.Name, &account.Email, &account.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

func (r *repo) GetAccountByID(ctx context.Context, id uint64) (*model.Account, error) {
	query := `SELECT id, name, email FROM accounts where id=$1`

	var account model.Account
	err := r.db.QueryRowContext(ctx, query, id).Scan(&account.ID, &account.Name, &account.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *repo) ListAccounts(ctx context.Context, skip, take uint64) ([]model.Account, error) {
	query := `SELECT id, name, email FROM accounts LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, take, skip)
	if err != nil {
		return []model.Account{}, err
	}
	defer rows.Close()

	accounts := []model.Account{}
	for rows.Next() {
		var account model.Account

		err := rows.Scan(&account.ID, &account.Name, &account.Email)
		if err != nil {
			return accounts, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *repo) PutTransaction(ctx context.Context, t model.Transaction) (*model.Transaction, error) {
	query := `INSERT INTO transactions (account_id, amount_cents, kind, description) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, t.AccountID, t.AmountCents, t.Kind, t.Description).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repo) ListTransactions(ctx context.Context, accountID uint64, skip, take uint64) ([]model.Transaction, error) {
	query := `SELECT id, account_id, amount_cents, kind, description, created_at FROM transactions WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, accountID, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.AccountID, &t.AmountCents, &t.Kind, &t.Description, &t.CreatedAt); err != nil {
			return list, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *repo) PutActivityLog(ctx context.Context, a model.ActivityLog) (*model.ActivityLog, error) {
	query := `INSERT INTO activity_log (account_id, action, ip_address) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, a.AccountID, a.Action, a.IPAddress).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repo) ListActivityLog(ctx context.Context, accountID uint64, skip, take uint64) ([]model.ActivityLog, error) {
	query := `SELECT id, account_id, action, ip_address, created_at FROM activity_log WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, accountID, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.ActivityLog
	for rows.Next() {
		var a model.ActivityLog
		if err := rows.Scan(&a.ID, &a.AccountID, &a.Action, &a.IPAddress, &a.CreatedAt); err != nil {
			return list, err
		}
		list = append(list, a)
	}
	return list, nil
}
