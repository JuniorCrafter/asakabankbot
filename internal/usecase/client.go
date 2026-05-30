package usecase

import (
	"asakabankbot/internal/domain"
	"asakabankbot/internal/repository"
)

// ClientUseCase определяет бизнес-логику для клиентов
type ClientUseCase interface {
	RegisterUser(tgID int64, username string) error
	UpdateUserState(tgID int64, state string) error
	GetProfile(tgID int64) (*domain.User, error)
}

type clientUseCase struct {
	userRepo *repository.UserRepository
}

func NewClientUseCase(userRepo *repository.UserRepository) ClientUseCase {
	return &clientUseCase{
		userRepo: userRepo,
	}
}

func (u *clientUseCase) RegisterUser(tgID int64, username string) error {
	return u.userRepo.CreateUser(tgID, username)
}

func (u *clientUseCase) UpdateUserState(tgID int64, state string) error {
	return u.userRepo.UpdateBotState(tgID, state)
}

func (u *clientUseCase) GetProfile(tgID int64) (*domain.User, error) {
	user, err := u.userRepo.GetUserByTelegramID(tgID)
	return user, err
}
