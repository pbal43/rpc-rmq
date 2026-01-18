package server

import (
	"client/internal"
	"context"
	"fmt"
	"log"
	pb "multiply-proto/multiplypb"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	rmqURL    string
	queueName string
	conn      *amqp.Connection
	ch        *amqp.Channel
}

func NewClient(cfg internal.Config) *Client {
	return &Client{
		rmqURL:    cfg.RmqURL,
		queueName: cfg.RpcQueueName,
	}
}

func (s *Client) Run() {
	conn, err := amqp.Dial(s.rmqURL)
	if err != nil {
		log.Fatal("Failed to connect: ", err)
	}
	s.conn = conn

	fmt.Println("Connected to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: ", err)
	}
	s.ch = ch

	fmt.Println("Channel opened")
}

func (c *Client) Call(number float32, timeout time.Duration) (float32, error) {
	replyQueue, err := c.ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return 0, err
	}

	msgs, err := c.ch.Consume(
		replyQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return 0, err
	}

	corrID := fmt.Sprintf("%s", uuid.New().String())

	req := &pb.MultiplyRequest{Number: number}
	body, _ := proto.Marshal(req)

	err = c.ch.Publish(
		"",
		c.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/protobuf",
			Body:          body,
			ReplyTo:       replyQueue.Name,
			CorrelationId: corrID,
		},
	)

	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case msg := <-msgs:
			if msg.CorrelationId == corrID {
				var resp pb.MultiplyResponse
				if err = proto.Unmarshal(msg.Body, &resp); err != nil {
					msg.Nack(false, false)
					return 0, err
				}
				msg.Ack(false)
				return resp.Message, nil
			} else {
				msg.Nack(false, true)
			}
		case <-ctx.Done():
			return 0, fmt.Errorf("timeout waiting for response")
		}
	}
}

func (s *Client) Stop() {
	if s.ch != nil {
		s.ch.Close()
	}

	if s.conn != nil {
		s.conn.Close()
	}
}
