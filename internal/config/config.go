package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	DBUrl          string
	RabbitMQUrl    string
	OperatorChatID int64
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не найден, используются системные переменные")
	}

	// 1. Ищем готовую ссылку (приоритет для Docker)
	dbUrl := os.Getenv("DB_URL")

	// 2. Если готовой ссылки нет, собираем по кусочкам (для локальной разработки)
	if dbUrl == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		dbUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPass, dbHost, dbPort, dbName)
	}

	// 3. Проверяем оба варианта имени для RabbitMQ
	rmqUrl := os.Getenv("RABBITMQ_URL")
	if rmqUrl == "" {
		rmqUrl = os.Getenv("RMQ_URL")
	}

	opID, _ := strconv.ParseInt(os.Getenv("OPERATOR_CHAT_ID"), 10, 64)

	return &Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		DBUrl:          dbUrl,
		RabbitMQUrl:    rmqUrl,
		OperatorChatID: opID,
	}
}
