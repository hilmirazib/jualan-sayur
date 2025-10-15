package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	VHost    string `json:"vhost"`
}

func (cfg Config) ConnectionRabbitMQ() (*amqp.Channel, error) {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.VHost)

	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Error().Err(err).Msg("[ConnectionRabbitMQ] Failed to connect to RabbitMQ")
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("[ConnectionRabbitMQ] Failed to open channel")
		conn.Close()
		return nil, err
	}

	// Declare queue for email verification
	_, err = channel.QueueDeclare(
		"email_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Error().Err(err).Msg("[ConnectionRabbitMQ] Failed to declare queue")
		channel.Close()
		conn.Close()
		return nil, err
	}

	log.Info().Msg("[ConnectionRabbitMQ] Successfully connected to RabbitMQ")
	return channel, nil
}
