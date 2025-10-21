package repository

import (
	"context"
	"strconv"
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

	// Parse lat/lng from string to float64
	lat, lng, err := u.parseLatLng(modelUser.Lat, modelUser.Lng)
	if err != nil {
		log.Error().Err(err).Str("lat", modelUser.Lat).Str("lng", modelUser.Lng).Msg("[UserRepository-GetUserByEmail] Failed to parse lat/lng")
		return nil, err
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      email,
		Password:   modelUser.Password,
		RoleName:   roleName,
		Address:    modelUser.Address,
		Lat:        lat,
		Lng:        lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func (u *UserRepository) CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error) {
	// Format lat/lng from float64 to string for database
	latStr, lngStr := u.formatLatLng(user.Lat, user.Lng)

	modelUser := &model.User{
		Name:       user.Name,
		Email:      user.Email,
		Password:   user.Password,
		Address:    user.Address,
		Lat:        latStr,
		Lng:        lngStr,
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

	// Parse lat/lng back to float64 for entity
	lat, lng, err := u.parseLatLng(modelUser.Lat, modelUser.Lng)
	if err != nil {
		log.Error().Err(err).Str("lat", modelUser.Lat).Str("lng", modelUser.Lng).Msg("[UserRepository-CreateUser] Failed to parse lat/lng")
		return nil, err
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		Password:   modelUser.Password,
		RoleName:   customerRole.Name,
		Address:    modelUser.Address,
		Lat:        lat,
		Lng:        lng,
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

func (u *UserRepository) UpdateUserVerificationStatus(ctx context.Context, userID int64, isVerified bool) error {
	if err := u.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("is_verified", isVerified).Error; err != nil {
		log.Error().Err(err).Int64("user_id", userID).Bool("is_verified", isVerified).Msg("[UserRepository-UpdateUserVerificationStatus] Failed to update user verification status")
		return err
	}

	log.Info().Int64("user_id", userID).Bool("is_verified", isVerified).Msg("[UserRepository-UpdateUserVerificationStatus] User verification status updated successfully")
	return nil
}

// GetUserByEmailIncludingUnverified implements UserRepositoryInterface.
func (u *UserRepository) GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}
	if err := u.db.Where("id = ? AND is_verified = ?", userID, true).Preload("Roles").First(&modelUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Int64("user_id", userID).Msg("[UserRepository-GetUserByID] User not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Int64("user_id", userID).Msg("[UserRepository-GetUserByID] Failed to get user by ID")
		return nil, err
	}

	var roleName string
	if len(modelUser.Roles) > 0 {
		roleName = modelUser.Roles[0].Name
	} else {
		roleName = "user" // Default role
	}

	// Parse lat/lng from string to float64
	lat, lng, err := u.parseLatLng(modelUser.Lat, modelUser.Lng)
	if err != nil {
		log.Error().Err(err).Str("lat", modelUser.Lat).Str("lng", modelUser.Lng).Msg("[UserRepository-GetUserByID] Failed to parse lat/lng")
		return nil, err
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		Password:   modelUser.Password,
		RoleName:   roleName,
		Address:    modelUser.Address,
		Lat:        lat,
		Lng:        lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func (u *UserRepository) GetUserByEmailIncludingUnverified(ctx context.Context, email string) (*entity.UserEntity, error) {
	modelUser := model.User{}
	if err := u.db.Where("email = ?", email).Preload("Roles").First(&modelUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Str("email", email).Msg("[UserRepository-GetUserByEmailIncludingUnverified] User not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Str("email", email).Msg("[UserRepository-GetUserByEmailIncludingUnverified] Failed to get user by email")
		return nil, err
	}

	// Check if user has roles
	var roleName string
	if len(modelUser.Roles) > 0 {
		roleName = modelUser.Roles[0].Name
	} else {
		roleName = "user" // Default role
	}

	// Parse lat/lng from string to float64
	lat, lng, err := u.parseLatLng(modelUser.Lat, modelUser.Lng)
	if err != nil {
		log.Error().Err(err).Str("lat", modelUser.Lat).Str("lng", modelUser.Lng).Msg("[UserRepository-GetUserByEmailIncludingUnverified] Failed to parse lat/lng")
		return nil, err
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      email,
		Password:   modelUser.Password,
		RoleName:   roleName,
		Address:    modelUser.Address,
		Lat:        lat,
		Lng:        lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func (u *UserRepository) UpdateUserPassword(ctx context.Context, userID int64, hashedPassword string) error {
	if err := u.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[UserRepository-UpdateUserPassword] Failed to update user password")
		return err
	}

	log.Info().Int64("user_id", userID).Msg("[UserRepository-UpdateUserPassword] User password updated successfully")
	return nil
}

func (u *UserRepository) UpdateUserPhoto(ctx context.Context, userID int64, photoURL string) error {
	if err := u.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("photo", photoURL).Error; err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("photo_url", photoURL).Msg("[UserRepository-UpdateUserPhoto] Failed to update user photo")
		return err
	}

	log.Info().Int64("user_id", userID).Str("photo_url", photoURL).Msg("[UserRepository-UpdateUserPhoto] User photo updated successfully")
	return nil
}

func (u *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, name, email, phone, address string, lat, lng float64, photo string) error {
	// Format lat/lng from float64 to string for database
	latStr, lngStr := u.formatLatLng(lat, lng)

	updates := map[string]interface{}{
		"name":    name,
		"email":   email,
		"phone":   phone,
		"address": address,
		"lat":     latStr,
		"lng":     lngStr,
		"photo":   photo,
	}

	if err := u.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("email", email).Msg("[UserRepository-UpdateUserProfile] Failed to update user profile")
		return err
	}

	log.Info().Int64("user_id", userID).Str("email", email).Msg("[UserRepository-UpdateUserProfile] User profile updated successfully")
	return nil
}

// Helper functions for lat/lng conversion
func (u *UserRepository) parseLatLng(latStr, lngStr string) (float64, float64, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, err
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lng, nil
}

func (u *UserRepository) formatLatLng(lat, lng float64) (string, string) {
	return strconv.FormatFloat(lat, 'f', -1, 64), strconv.FormatFloat(lng, 'f', -1, 64)
}

func NewUserRepository(db *gorm.DB) port.UserRepositoryInterface {
	return &UserRepository{db: db}
}
