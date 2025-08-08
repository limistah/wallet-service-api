package usecases

import (
	"errors"

	"github.com/limistah/wallet-service/internal/models"
	"github.com/limistah/wallet-service/internal/repositories"
	"github.com/limistah/wallet-service/internal/utils"
	"gorm.io/gorm"
)

type userUseCase struct {
	repos *repositories.Repositories
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(repos *repositories.Repositories) UserUseCase {
	return &userUseCase{repos: repos}
}

func (uc *userUseCase) CreateUser(user *models.User) (*models.User, error) {
	if err := utils.ValidateStruct(user); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := uc.repos.User.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Use transaction to ensure data consistency
	var createdUser *models.User
	err = uc.repos.DB.Transaction(func(tx *gorm.DB) error {
		// Create user within transaction
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		createdUser = user

		// Create default wallet for the user within the same transaction
		wallet := &models.Wallet{
			UserID:   user.ID,
			Currency: "USD",
			Status:   models.WalletStatusActive,
		}

		if err := tx.Create(wallet).Error; err != nil {
			return errors.New("failed to create user wallet")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (uc *userUseCase) GetUser(id uint) (*models.User, error) {
	return uc.repos.User.GetByID(id)
}

func (uc *userUseCase) GetUserByID(id uint) (*models.User, error) {
	return uc.repos.User.GetByID(id)
}

func (uc *userUseCase) GetUserByEmail(email string) (*models.User, error) {
	return uc.repos.User.GetByEmail(email)
}

func (uc *userUseCase) UpdateUser(id uint, updatedUser *models.User) (*models.User, error) {
	user, err := uc.repos.User.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update only non-empty fields
	if updatedUser.Name != "" {
		user.Name = updatedUser.Name
	}
	if updatedUser.Age != 0 {
		user.Age = updatedUser.Age
	}
	// Don't allow email updates through this method for security
	// Password updates should go through ChangePassword method
	if updatedUser.Password != "" {
		user.Password = updatedUser.Password
	}

	err = uc.repos.User.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) DeleteUser(id uint) error {
	return uc.repos.User.Delete(id)
}

func (uc *userUseCase) ListUsers(page, pageSize int) ([]models.User, error) {
	offset := (page - 1) * pageSize
	return uc.repos.User.List(offset, pageSize)
}
