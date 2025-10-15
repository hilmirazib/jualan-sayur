package repository

import (
	"context"
	"time"
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

func (r *VerificationTokenRepository) GetVerificationToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error) {
	modelToken := &model.VerificationToken{}
	if err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(modelToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &entity.VerificationTokenEntity{
		ID:        modelToken.ID,
		UserID:    modelToken.UserID,
		Token:     modelToken.Token,
		TokenType: modelToken.TokenType,
		ExpiresAt: modelToken.ExpiresAt,
	}, nil
}

func (r *VerificationTokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.VerificationToken{}).Error
}
