package repository

import (
	"context"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"
	"user-service/internal/core/port"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

// GetUserByEmail implements UserRepositoryInterface.
func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	modelUser := model.User{}
	if err := u.db.Where("email = ? AND is_verified = ?", email, true).Preload("Roles").First(&modelUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Str("email", email).Msg("[UserRepository-GetUserByEmail] User not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Str("email", email).Msg("[UserRepository-GetUserByEmail] Failed to get user by email")
		return nil, err
	}

	// Check if user has roles
	var roleName string
	if len(modelUser.Roles) > 0 {
		roleName = modelUser.Roles[0].Name
	} else {
		roleName = "user" // Default role
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      email,
		Password:   modelUser.Password,
		RoleName:   roleName,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func NewUserRepository(db *gorm.DB) port.UserRepositoryInterface {
	return &UserRepository{db: db}
}
