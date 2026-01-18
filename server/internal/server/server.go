package server

import (
	"context"
	"fmt"
	"log"
	"server/internal"

	pb "multiply-proto/multiplypb"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	rmqURL    string
	queueName string
	conn      *amqp.Connection
	ch        *amqp.Channel
}

func NewServer(cfg internal.Config) *Server {
	return &Server{rmqURL: cfg.RmqURL, queueName: cfg.RpcQueueName}
}

func (s *Server) Run() {
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

	queue, err := ch.QueueDeclare(
		s.queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal("Failed to declare a queue: ", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Printf("Failed to set QoS: %s", err)
	}

	fmt.Printf("Queue declared: %s\n", queue.Name)
}

func (s *Server) StartConsume() {
	msgCount := 0

	msgs, err := s.ch.Consume(
		s.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal("Failed to register a consumer: ", err)
	}

	log.Println("Consuming started")

	for msg := range msgs {
		msgCount++

		var req pb.MultiplyRequest
		if err = proto.Unmarshal(msg.Body, &req); err != nil {
			msg.Nack(false, false)
			continue
		}

		resp := &pb.MultiplyResponse{
			Success:   true,
			Message:   req.Number * 2,
			RequestId: int64(msgCount),
			TimeStamp: msg.Timestamp.Format("2006-01-02 15:04:05"),
		}

		body, _ := proto.Marshal(resp)

		err = s.ch.PublishWithContext(
			context.Background(),
			"",
			msg.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/protobuf",
				Body:          body,
				CorrelationId: msg.CorrelationId,
			},
		)
		if err != nil {
			log.Printf("Failed to publish a message: %s", err)
			msg.Nack(false, false)
			continue
		}

		msg.Ack(false)
	}
}

func (s *Server) Stop() {
	if s.ch != nil {
		s.ch.Close()
	}

	if s.conn != nil {
		s.conn.Close()
	}
}
