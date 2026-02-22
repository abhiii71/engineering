package internal

import (
	"context"
	"database/sql"

	model "github.com/abhiii71/orderStream/account-service/models"
)

type AccountRepository interface {
	Close() error
	PutAccount(ctx context.Context, a model.Account) (*model.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*model.Account, error)
	GetAccountByID(ctx context.Context, id uint64) (*model.Account, error)
	ListAccounts(ctx context.Context, skip, take uint64) ([]model.Account, error)
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
