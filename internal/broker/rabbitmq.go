package broker

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectRabbitMQ(url string) *amqp.Connection {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Ошибка при подключении к RabbitMQ: %v", err)
	}

	log.Println("Успешное подключение к RabbitMQ!")
	return conn
}
