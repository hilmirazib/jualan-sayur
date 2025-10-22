package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type RoleRepositoryInterface interface {
	GetAllRoles(ctx context.Context, search string) ([]entity.RoleEntity, error)
}
