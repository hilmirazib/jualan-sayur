package repository

import (
	"context"
	"fmt"
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

func (r *RoleRepository) CreateRole(ctx context.Context, role *entity.RoleEntity) (*entity.RoleEntity, error) {
	roleModel := &model.Role{
		Name: role.Name,
	}

	if err := r.db.WithContext(ctx).Create(roleModel).Error; err != nil {
		log.Error().Err(err).Str("role_name", role.Name).Msg("[RoleRepository-CreateRole] Failed to create role")
		return nil, err
	}

	// Convert back to entity
	createdRole := &entity.RoleEntity{
		ID:        roleModel.ID,
		Name:      roleModel.Name,
		CreatedAt: roleModel.CreatedAt,
		UpdatedAt: roleModel.UpdatedAt,
	}

	log.Info().Int64("role_id", createdRole.ID).Str("role_name", createdRole.Name).Msg("[RoleRepository-CreateRole] Role created successfully")
	return createdRole, nil
}

func (r *RoleRepository) UpdateRole(ctx context.Context, id int64, role *entity.RoleEntity) (*entity.RoleEntity, error) {
	var existingRole model.Role
	if err := r.db.WithContext(ctx).First(&existingRole, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Int64("role_id", id).Msg("[RoleRepository-UpdateRole] Role not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleRepository-UpdateRole] Failed to find role")
		return nil, err
	}

	// Update fields
	existingRole.Name = role.Name

	if err := r.db.WithContext(ctx).Save(&existingRole).Error; err != nil {
		log.Error().Err(err).Int64("role_id", id).Str("role_name", role.Name).Msg("[RoleRepository-UpdateRole] Failed to update role")
		return nil, err
	}

	// Convert back to entity
	updatedRole := &entity.RoleEntity{
		ID:        existingRole.ID,
		Name:      existingRole.Name,
		CreatedAt: existingRole.CreatedAt,
		UpdatedAt: existingRole.UpdatedAt,
	}

	log.Info().Int64("role_id", id).Str("role_name", role.Name).Msg("[RoleRepository-UpdateRole] Role updated successfully")
	return updatedRole, nil
}

func (r *RoleRepository) DeleteRole(ctx context.Context, id int64) error {
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Int64("role_id", id).Msg("[RoleRepository-DeleteRole] Role not found")
			return gorm.ErrRecordNotFound
		}
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleRepository-DeleteRole] Failed to find role")
		return err
	}

	// Check if role has associated users using a join query
	var userCount int64
	if err := r.db.WithContext(ctx).Table("user_role").
		Where("role_id = ?", id).
		Count(&userCount).Error; err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleRepository-DeleteRole] Failed to check associated users")
		return err
	}

	if userCount > 0 {
		log.Warn().Int64("role_id", id).Int64("user_count", userCount).Msg("[RoleRepository-DeleteRole] Cannot delete role with associated users")
		return fmt.Errorf("cannot delete role that is currently assigned to users")
	}

	// Soft delete the role
	if err := r.db.WithContext(ctx).Delete(&role).Error; err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleRepository-DeleteRole] Failed to delete role")
		return err
	}

	log.Info().Int64("role_id", id).Str("role_name", role.Name).Msg("[RoleRepository-DeleteRole] Role deleted successfully")
	return nil
}

func NewRoleRepository(db *gorm.DB) port.RoleRepositoryInterface {
	return &RoleRepository{db: db}
}
