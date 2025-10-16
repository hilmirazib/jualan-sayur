package consumer

import (
	"context"
	"encoding/json"
	"notification-service/config"
	"notification-service/internal/core/port"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type EmailMessage struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailConsumer struct {
	config       *config.Config
	emailService port.EmailServiceInterface
	channel      *amqp.Channel
}

func NewEmailConsumer(cfg *config.Config, emailService port.EmailServiceInterface, channel *amqp.Channel) *EmailConsumer {
	return &EmailConsumer{
		config:       cfg,
		emailService: emailService,
		channel:      channel,
	}
}

func (c *EmailConsumer) StartConsuming(ctx context.Context) error {
	// Declare queue (same as publisher)
	queue, err := c.channel.QueueDeclare(
		"email_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Error().Err(err).Msg("[EmailConsumer-StartConsuming] Failed to declare queue")
		return err
	}

	// Start consuming messages
	msgs, err := c.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Error().Err(err).Msg("[EmailConsumer-StartConsuming] Failed to register consumer")
		return err
	}

	log.Info().Msg("[EmailConsumer-StartConsuming] Started consuming email messages")

	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("[EmailConsumer-StartConsuming] Stopping consumer")
				return
			case msg := <-msgs:
				c.processMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *EmailConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var emailMsg EmailMessage
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		log.Error().Err(err).Msg("[EmailConsumer-processMessage] Failed to unmarshal message")
		msg.Nack(false, false) // Don't requeue
		return
	}

	log.Info().Str("email", emailMsg.Email).Str("type", emailMsg.Type).Msg("[EmailConsumer-processMessage] Processing email message")

	// Send email using email service
	if err := c.emailService.SendEmail(ctx, emailMsg.Email, emailMsg.Subject, emailMsg.Body); err != nil {
		log.Error().Err(err).Str("email", emailMsg.Email).Msg("[EmailConsumer-processMessage] Failed to send email")
		msg.Nack(false, true) // Requeue for retry
		return
	}

	log.Info().Str("email", emailMsg.Email).Msg("[EmailConsumer-processMessage] Email sent successfully")

	// Acknowledge message
	msg.Ack(false)
}
