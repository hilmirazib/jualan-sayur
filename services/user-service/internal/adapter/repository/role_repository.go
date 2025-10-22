package repository

import (
	"context"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"
	"user-service/internal/core/port"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func (r *RoleRepository) GetAllRoles(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	var roles []model.Role
	query := r.db.WithContext(ctx)

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Find(&roles).Error; err != nil {
		log.Error().Err(err).Str("search", search).Msg("[RoleRepository-GetAllRoles] Failed to get roles")
		return nil, err
	}

	var roleEntities []entity.RoleEntity
	for _, role := range roles {
		roleEntities = append(roleEntities, entity.RoleEntity{
			ID:        role.ID,
			Name:      role.Name,
			CreatedAt: role.CreatedAt,
			UpdatedAt: role.UpdatedAt,
			DeletedAt: role.DeletedAt,
		})
	}

	log.Info().Int("count", len(roleEntities)).Str("search", search).Msg("[RoleRepository-GetAllRoles] Roles retrieved successfully")
	return roleEntities, nil
}

func (r *RoleRepository) GetRoleByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Preload("Users").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Int64("role_id", id).Msg("[RoleRepository-GetRoleByID] Role not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleRepository-GetRoleByID] Failed to get role by ID")
		return nil, err
	}

	// Convert users to entity format
	var userEntities []entity.UserEntity
	for _, user := range role.Users {
		userEntities = append(userEntities, entity.UserEntity{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	roleEntity := &entity.RoleEntity{
		ID:        role.ID,
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
		DeletedAt: role.DeletedAt,
		Users:     userEntities,
	}

	log.Info().Int64("role_id", id).Int("users_count", len(userEntities)).Msg("[RoleRepository-GetRoleByID] Role retrieved successfully")
	return roleEntity, nil
}

func NewRoleRepository(db *gorm.DB) port.RoleRepositoryInterface {
	return &RoleRepository{db: db}
}
