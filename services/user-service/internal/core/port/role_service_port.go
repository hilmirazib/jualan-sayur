package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type RoleServiceInterface interface {
	GetAllRoles(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetRoleByID(ctx context.Context, id int64) (*entity.RoleEntity, error)
	CreateRole(ctx context.Context, name string) (*entity.RoleEntity, error)
	UpdateRole(ctx context.Context, id int64, name string) (*entity.RoleEntity, error)
	DeleteRole(ctx context.Context, id int64) error
}
