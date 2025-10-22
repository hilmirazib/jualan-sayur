package service

import (
	"context"
	"fmt"
	"strings"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"

	"github.com/rs/zerolog/log"
)

type RoleService struct {
	roleRepo port.RoleRepositoryInterface
}

func (s *RoleService) GetAllRoles(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	roles, err := s.roleRepo.GetAllRoles(ctx, search)
	if err != nil {
		log.Error().Err(err).Str("search", search).Msg("[RoleService-GetAllRoles] Failed to get roles")
		return nil, err
	}

	log.Info().Int("count", len(roles)).Str("search", search).Msg("[RoleService-GetAllRoles] Roles retrieved successfully")
	return roles, nil
}

func (s *RoleService) GetRoleByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	role, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleService-GetRoleByID] Failed to get role by ID")
		return nil, err
	}

	log.Info().Int64("role_id", id).Msg("[RoleService-GetRoleByID] Role retrieved successfully")
	return role, nil
}

func (s *RoleService) CreateRole(ctx context.Context, name string) (*entity.RoleEntity, error) {
	// Validate input
	if name == "" {
		log.Warn().Msg("[RoleService-CreateRole] Role name cannot be empty")
		return nil, fmt.Errorf("role name cannot be empty")
	}

	// Trim whitespace
	name = strings.TrimSpace(name)
	if name == "" {
		log.Warn().Msg("[RoleService-CreateRole] Role name cannot be only whitespace")
		return nil, fmt.Errorf("role name cannot be empty")
	}

	// Check length
	if len(name) < 2 || len(name) > 50 {
		log.Warn().Int("name_length", len(name)).Msg("[RoleService-CreateRole] Role name length invalid")
		return nil, fmt.Errorf("role name must be between 2 and 50 characters")
	}

	// Check if role already exists
	existingRoles, err := s.roleRepo.GetAllRoles(ctx, "")
	if err != nil {
		log.Error().Err(err).Msg("[RoleService-CreateRole] Failed to check existing roles")
		return nil, err
	}

	for _, role := range existingRoles {
		if strings.EqualFold(role.Name, name) {
			log.Warn().Str("role_name", name).Msg("[RoleService-CreateRole] Role name already exists")
			return nil, fmt.Errorf("role with name '%s' already exists", name)
		}
	}

	// Create role entity
	roleEntity := &entity.RoleEntity{
		Name: name,
	}

	// Create role in repository
	createdRole, err := s.roleRepo.CreateRole(ctx, roleEntity)
	if err != nil {
		log.Error().Err(err).Str("role_name", name).Msg("[RoleService-CreateRole] Failed to create role")
		return nil, err
	}

	log.Info().Int64("role_id", createdRole.ID).Str("role_name", createdRole.Name).Msg("[RoleService-CreateRole] Role created successfully")
	return createdRole, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id int64, name string) (*entity.RoleEntity, error) {
	// Validate input
	if name == "" {
		log.Warn().Int64("role_id", id).Msg("[RoleService-UpdateRole] Role name cannot be empty")
		return nil, fmt.Errorf("role name cannot be empty")
	}

	// Trim whitespace
	name = strings.TrimSpace(name)
	if name == "" {
		log.Warn().Int64("role_id", id).Msg("[RoleService-UpdateRole] Role name cannot be only whitespace")
		return nil, fmt.Errorf("role name cannot be empty")
	}

	// Check length
	if len(name) < 2 || len(name) > 50 {
		log.Warn().Int64("role_id", id).Int("name_length", len(name)).Msg("[RoleService-UpdateRole] Role name length invalid")
		return nil, fmt.Errorf("role name must be between 2 and 50 characters")
	}

	// Check if role exists
	existingRole, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleService-UpdateRole] Failed to get existing role")
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	// Check if another role with the same name already exists (excluding current role)
	allRoles, err := s.roleRepo.GetAllRoles(ctx, "")
	if err != nil {
		log.Error().Err(err).Msg("[RoleService-UpdateRole] Failed to check existing roles")
		return nil, err
	}

	for _, role := range allRoles {
		if role.ID != id && strings.EqualFold(role.Name, name) {
			log.Warn().Int64("role_id", id).Str("role_name", name).Msg("[RoleService-UpdateRole] Role name already exists")
			return nil, fmt.Errorf("role with name '%s' already exists", name)
		}
	}

	// Create updated role entity
	updatedRoleEntity := &entity.RoleEntity{
		Name: name,
	}

	// Update role in repository
	updatedRole, err := s.roleRepo.UpdateRole(ctx, id, updatedRoleEntity)
	if err != nil {
		log.Error().Err(err).Int64("role_id", id).Str("role_name", name).Msg("[RoleService-UpdateRole] Failed to update role")
		return nil, err
	}

	log.Info().Int64("role_id", id).Str("old_name", existingRole.Name).Str("new_name", updatedRole.Name).Msg("[RoleService-UpdateRole] Role updated successfully")
	return updatedRole, nil
}

func NewRoleService(roleRepo port.RoleRepositoryInterface) port.RoleServiceInterface {
	return &RoleService{
		roleRepo: roleRepo,
	}
}
