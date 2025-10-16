package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"user-service/internal/core/port"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type EmailPublisher struct {
	channel *amqp.Channel
}

type EmailVerificationMessage struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewEmailPublisher(channel *amqp.Channel) port.EmailInterface {
	return &EmailPublisher{
		channel: channel,
	}
}

func (p *EmailPublisher) SendVerificationEmail(ctx context.Context, email, token string) error {
	// Extract name from email (before @) or use default
	name := "User"
	if atIndex := strings.Index(email, "@"); atIndex > 0 {
		name = email[:atIndex]
		// Capitalize first letter
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
		}
	}

	verificationLink := "http://localhost:8080/api/v1/auth/verify?token=" + token

	message := EmailVerificationMessage{
		Email:   email,
		Token:   token,
		Type:    "email_verification",
		Name:    name,
		Subject: "Verify Your Account",
		Body: fmt.Sprintf(`Hi %s,

Please click this link to verify your account:
%s

Link expires in 24 hours.

If you didn't create an account, please ignore this email.

Best regards,
Your App Team`, name, verificationLink),
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("[EmailPublisher-SendVerificationEmail] Failed to marshal message")
		return err
	}

	err = p.channel.Publish(
		"",            // exchange
		"email_queue", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("[EmailPublisher-SendVerificationEmail] Failed to publish message")
		return err
	}

	log.Info().Str("email", email).Msg("[EmailPublisher-SendVerificationEmail] Verification email sent to queue")
	return nil
}

func (p *EmailPublisher) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	// Extract name from email (before @) or use default
	name := "User"
	if atIndex := strings.Index(email, "@"); atIndex > 0 {
		name = email[:atIndex]
		// Capitalize first letter
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
		}
	}

	resetLink := "http://localhost:8080/api/v1/auth/reset-password?token=" + token

	message := EmailVerificationMessage{
		Email:   email,
		Token:   token,
		Type:    "password_reset",
		Name:    name,
		Subject: "Reset Your Password",
		Body: fmt.Sprintf(`Hi %s,

You requested to reset your password. Please click this link to reset your password:
%s

Link expires in 1 hour.

If you didn't request this, please ignore this email.

Best regards,
Your App Team`, name, resetLink),
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("[EmailPublisher-SendPasswordResetEmail] Failed to marshal message")
		return err
	}

	err = p.channel.Publish(
		"",            // exchange
		"email_queue", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("[EmailPublisher-SendPasswordResetEmail] Failed to publish message")
		return err
	}

	log.Info().Str("email", email).Msg("[EmailPublisher-SendPasswordResetEmail] Password reset email sent to queue")
	return nil
}
