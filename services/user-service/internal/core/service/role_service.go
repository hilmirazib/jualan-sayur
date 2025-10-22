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

func NewRoleService(roleRepo port.RoleRepositoryInterface) port.RoleServiceInterface {
	return &RoleService{
		roleRepo: roleRepo,
	}
}
