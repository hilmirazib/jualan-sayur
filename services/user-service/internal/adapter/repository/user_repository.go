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

func (u *UserRepository) CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error) {
	modelUser := &model.User{
		Name:       user.Name,
		Email:      user.Email,
		Password:   user.Password,
		Address:    user.Address,
		Lat:        user.Lat,
		Lng:        user.Lng,
		Phone:      user.Phone,
		Photo:      user.Photo,
		IsVerified: user.IsVerified,
	}

	if err := u.db.WithContext(ctx).Create(modelUser).Error; err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("[UserRepository-CreateUser] Failed to create user")
		return nil, err
	}

	// Assign default role "Customer"
	customerRole := &model.Role{}
	if err := u.db.Where("name = ?", "Customer").First(customerRole).Error; err != nil {
		log.Error().Err(err).Msg("[UserRepository-CreateUser] Failed to find Customer role")
		return nil, err
	}

	if err := u.db.Model(modelUser).Association("Roles").Append(customerRole); err != nil {
		log.Error().Err(err).Int64("user_id", modelUser.ID).Msg("[UserRepository-CreateUser] Failed to assign role")
		return nil, err
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		Password:   modelUser.Password,
		RoleName:   customerRole.Name,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func (u *UserRepository) GetRoleByName(ctx context.Context, name string) (*entity.RoleEntity, error) {
	modelRole := &model.Role{}
	if err := u.db.Where("name = ?", name).First(modelRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Str("role_name", name).Msg("[UserRepository-GetRoleByName] Role not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Str("role_name", name).Msg("[UserRepository-GetRoleByName] Failed to get role by name")
		return nil, err
	}

	return &entity.RoleEntity{
		ID:        modelRole.ID,
		Name:      modelRole.Name,
		CreatedAt: modelRole.CreatedAt,
		UpdatedAt: modelRole.UpdatedAt,
		DeletedAt: modelRole.DeletedAt,
	}, nil
}

func NewUserRepository(db *gorm.DB) port.UserRepositoryInterface {
	return &UserRepository{db: db}
}
