package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"asakabankbot/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
}

func NewPublisher(conn *amqp.Connection) *Publisher {
	return &Publisher{conn: conn}
}

// SendTicket публикует заявку в очередь нужного отдела
func (p *Publisher) SendTicket(ticket domain.ChatTicket) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Называем очередь в зависимости от отдела, например: dep_1_queue
	queueName := fmt.Sprintf("dep_%d_queue", ticket.DepID)

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable (сохраняется при перезапуске RabbitMQ)
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(ticket)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})

	if err != nil {
		log.Printf("Ошибка отправки в RabbitMQ: %v", err)
		return err
	}

	log.Printf("Заявка сессии #%d успешно отправлена в очередь %s", ticket.SessionID, q.Name)
	return nil
}

// SendChatMessage отправляет текстовое сообщение от клиента или оператора в RabbitMQ
func (p *Publisher) SendChatMessage(msg domain.ChatMessage) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Очередь для всех сообщений внутри активных чатов
	q, err := ch.QueueDeclare(
		"active_chat_messages",
		true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(msg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})

	if err != nil {
		log.Printf("Ошибка отправки сообщения чата в RabbitMQ: %v", err)
		return err
	}
	return nil
}
