package service

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/entity"
	log "github.com/sirupsen/logrus"
)

type AccountService struct {
	aRepo entity.AccountRepository
}

func NewAccountService(aRepo entity.AccountRepository) *AccountService {
	return &AccountService{
		aRepo: aRepo,
	}
}

func (s *AccountService) AddAccount(ctx context.Context, provider entity.Provider, keyID string, secret string) (string, error) {
	log.Infof("add new account")
	account := entity.NewServiceAccount(entity.NewAccountID(), provider, keyID, secret)
	err := s.aRepo.Add(ctx, account)
	if err != nil {
		log.Errorf("failed to add account: %v", err.Error())
		return "", err
	}
	return account.ID, nil
}

func (s *AccountService) ListAccountByProvider(ctx context.Context, provider entity.Provider) ([]*entity.ServiceAccount, error) {
	log.Infof("search service accounts of provider %s", provider)
	accounts, err := s.aRepo.ListByProvider(ctx, provider)
	if err != nil {
		log.Errorf("failed to list accounts: %v", err.Error())
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	return accounts, nil
}

func (s *AccountService) GetAccountByID(ctx context.Context, accountID string) (*entity.ServiceAccount, error) {
	log.Infof("search account with id %s", accountID)
	account, err := s.aRepo.GetByID(ctx, accountID)

	if err != nil {
		log.Errorf("failed to find account: %v", err.Error())
		return nil, err
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account, nil
}
