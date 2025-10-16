package service

import (
	"context"
	"notification-service/config"
	"notification-service/internal/core/port"

	"github.com/rs/zerolog/log"
	gomail "gopkg.in/gomail.v2"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) port.EmailServiceInterface {
	return &EmailService{
		config: cfg,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	m := gomail.NewMessage()

	// Set email headers - use a proper from address for Mailtrap
	fromAddress := "noreply@mailtrap.io" // Use a standard Mailtrap from address
	m.SetHeader("From", fromAddress)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	// Set email body
	m.SetBody("text/html", body)

	// Create SMTP dialer
	d := gomail.NewDialer(s.config.SMTP.Host, s.config.SMTP.Port, s.config.SMTP.User, s.config.SMTP.Password)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		log.Error().Err(err).Str("to", to).Str("subject", subject).Msg("[EmailService-SendEmail] Failed to send email")
		return err
	}

	log.Info().Str("to", to).Str("subject", subject).Msg("[EmailService-SendEmail] Email sent successfully")
	return nil
}
