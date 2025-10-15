package message

import (
	"context"
	"encoding/json"
	"user-service/internal/core/port"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type EmailPublisher struct {
	channel *amqp.Channel
}

type EmailVerificationMessage struct {
	Email string `json:"email"`
	Token string `json:"token"`
	Type  string `json:"type"`
}

func NewEmailPublisher(channel *amqp.Channel) port.EmailInterface {
	return &EmailPublisher{
		channel: channel,
	}
}

func (p *EmailPublisher) SendVerificationEmail(ctx context.Context, email, token string) error {
	message := EmailVerificationMessage{
		Email: email,
		Token: token,
		Type:  "email_verification",
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("[EmailPublisher-SendVerificationEmail] Failed to marshal message")
		return err
	}

	err = p.channel.Publish(
		"",                // exchange
		"email_queue",     // routing key
		false,             // mandatory
		false,             // immediate
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
