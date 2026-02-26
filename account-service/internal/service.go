package internal

import (
	"context"
	"errors"

	model "github.com/abhiii71/engineering/account-service/models"
	"github.com/abhiii71/engineering/account-service/pkg/auth"
	"github.com/abhiii71/engineering/account-service/pkg/crypt"
)

type AccountService interface {
	Register(ctx context.Context, name, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetAccount(ctx context.Context, id uint64) (*model.Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]model.Account, error)
	RecordTransaction(ctx context.Context, accountID uint64, amountCents int64, kind, description string) (*model.Transaction, error)
	ListTransactions(ctx context.Context, accountID uint64, skip, take uint64) ([]model.Transaction, error)
	RecordActivity(ctx context.Context, accountID uint64, action, ipAddress string) (*model.ActivityLog, error)
	ListActivity(ctx context.Context, accountID uint64, skip, take uint64) ([]model.ActivityLog, error)
}

type service struct {
	repo AccountRepository
}

func NewService(r AccountRepository) AccountService {
	return &service{r}
}

func (s *service) Register(ctx context.Context, name, email, password string) (string, error) {
	account, err := s.repo.GetAccountByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if account != nil {
		return "", errors.New("account already exists")
	}

	hashedPassword, err := crypt.HashPassword(password)
	if err != nil {
		return "", err
	}

	acc := model.Account{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	account, err = s.repo.PutAccount(ctx, acc)
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(account.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	account, err := s.repo.GetAccountByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", errors.New("account not found")
	}

	err = crypt.VerifyPassword(password, account.Password)
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(account.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) GetAccount(ctx context.Context, id uint64) (*model.Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *service) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]model.Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return s.repo.ListAccounts(ctx, skip, take)
}

func (s *service) RecordTransaction(ctx context.Context, accountID uint64, amountCents int64, kind, description string) (*model.Transaction, error) {
	if kind != "credit" && kind != "debit" {
		return nil, errors.New("kind must be credit or debit")
	}
	t := model.Transaction{
		AccountID:   accountID,
		AmountCents: amountCents,
		Kind:        kind,
		Description: description,
	}
	return s.repo.PutTransaction(ctx, t)
}

func (s *service) ListTransactions(ctx context.Context, accountID uint64, skip, take uint64) ([]model.Transaction, error) {
	if take > 100 || take == 0 {
		take = 100
	}
	return s.repo.ListTransactions(ctx, accountID, skip, take)
}

func (s *service) RecordActivity(ctx context.Context, accountID uint64, action, ipAddress string) (*model.ActivityLog, error) {
	if action == "" {
		return nil, errors.New("action required")
	}
	a := model.ActivityLog{
		AccountID: accountID,
		Action:    action,
		IPAddress: ipAddress,
	}
	return s.repo.PutActivityLog(ctx, a)
}

func (s *service) ListActivity(ctx context.Context, accountID uint64, skip, take uint64) ([]model.ActivityLog, error) {
	if take > 100 || take == 0 {
		take = 100
	}
	return s.repo.ListActivityLog(ctx, accountID, skip, take)
}
