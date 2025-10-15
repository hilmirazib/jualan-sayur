package repository

import (
	"context"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"
	"user-service/internal/core/port"

	"gorm.io/gorm"
)

type VerificationTokenRepository struct {
	db *gorm.DB
}

func NewVerificationTokenRepository(db *gorm.DB) port.VerificationTokenInterface {
	return &VerificationTokenRepository{
		db: db,
	}
}

func (r *VerificationTokenRepository) CreateVerificationToken(ctx context.Context, token *entity.VerificationTokenEntity) error {
	model := &model.VerificationToken{
		UserID:    token.UserID,
		Token:     token.Token,
		TokenType: token.TokenType,
		ExpiresAt: token.ExpiresAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	token.ID = model.ID
	return nil
}
