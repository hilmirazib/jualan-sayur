package repository

import (
	"context"
	"time"
	"user-service/internal/core/domain/model"
	"user-service/internal/core/port"

	"gorm.io/gorm"
)

type BlacklistTokenRepository struct {
	db *gorm.DB
}

func NewBlacklistTokenRepository(db *gorm.DB) port.BlacklistTokenInterface {
	return &BlacklistTokenRepository{
		db: db,
	}
}

func (r *BlacklistTokenRepository) AddToBlacklist(ctx context.Context, tokenHash string, expiresAt int64) error {
	model := &model.BlacklistToken{
		TokenHash: tokenHash,
		ExpiresAt: time.Unix(expiresAt, 0),
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *BlacklistTokenRepository) IsTokenBlacklisted(ctx context.Context, tokenHash string) bool {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BlacklistToken{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error

	return err == nil && count > 0
}
