package service

import (
	"context"
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

func NewRoleService(roleRepo port.RoleRepositoryInterface) port.RoleServiceInterface {
	return &RoleService{
		roleRepo: roleRepo,
	}
}
