package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client wraps an AMQP connection/channel.
type Client struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

// New creates a new client and declares the queue.
func New(ctx context.Context, url string, queue string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_, err = ch.QueueDeclare(
		queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn, ch: ch, queue: queue}, nil
}

// Close closes channel and connection.
func (c *Client) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// Publish marshals payload to JSON and publishes to the queue.
func (c *Client) Publish(ctx context.Context, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.ch.PublishWithContext(ctx,
		"",
		c.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

// Consume starts consuming messages and passes each to handler.
func (c *Client) Consume(ctx context.Context, handler func(context.Context, []byte) error) error {
	deliveries, err := c.ch.Consume(
		c.queue,
		"",
		false, // auto-ack disabled
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				if err := handler(msgCtx, d.Body); err != nil {
					_ = d.Nack(false, true)
				} else {
					_ = d.Ack(false)
				}
				cancel()
			}
		}
	}()
	return nil
}
