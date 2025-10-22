package port

import (
	"context"
)

type EmailInterface interface {
	SendVerificationEmail(ctx context.Context, email, token string) error
	SendEmailChangeVerificationEmail(ctx context.Context, email, token string) error
	SendPasswordResetEmail(ctx context.Context, email, token string) error
}
