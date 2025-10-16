package port

import "context"

type EmailServiceInterface interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}
