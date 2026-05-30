package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"asakabankbot/internal/domain"
	"asakabankbot/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn        *amqp.Connection
	bot         *tgbotapi.BotAPI
	sessionRepo *repository.SessionRepository
	opRepo      *repository.OperatorRepository
}

func NewConsumer(conn *amqp.Connection, bot *tgbotapi.BotAPI, sessionRepo *repository.SessionRepository, opRepo *repository.OperatorRepository) *Consumer {
	return &Consumer{conn: conn, bot: bot, sessionRepo: sessionRepo, opRepo: opRepo}
}

func (c *Consumer) StartTicketConsumer(depID int) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Printf("Ошибка открытия канала RabbitMQ: %v", err)
		return
	}

	// ИСПРАВЛЕНИЕ: Ограничиваем получение одним сообщением за раз
	err = ch.Qos(
		1,     // prefetch count (сколько сообщений брать за раз)
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Printf("Ошибка настройки Qos: %v", err)
	}

	queueName := fmt.Sprintf("dep_%d_queue", depID)
	q, _ := ch.QueueDeclare(queueName, true, false, false, false, nil)

	// ИСПРАВЛЕНИЕ: autoAck = false (третий аргумент с конца)
	msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)

	go func() {
		for d := range msgs {
			var ticket domain.ChatTicket
			json.Unmarshal(d.Body, &ticket)

			if !c.sessionRepo.IsSessionActive(ticket.SessionID) {
				log.Printf("Заявка #%d отменена клиентом. Удаляем из RabbitMQ.", ticket.SessionID)
				d.Ack(false) // Уничтожаем сообщение навсегда
				continue
			}

			// Ищем ONLINE операторов нужного отдела
			operators, err := c.opRepo.GetOnlineOperatorsByDepartment(ticket.DepID)

			if err != nil || len(operators) == 0 {
				log.Printf("Внимание: для заявки #%d (Отдел %d) нет свободных операторов. Заявка возвращена в очередь.", ticket.SessionID, ticket.DepID)

				// Ждем 5 секунд, чтобы не перегружать CPU циклическими проверками
				time.Sleep(5 * time.Second)

				// Nack (Negative Acknowledge) возвращает сообщение в очередь (requeue = true)
				d.Nack(false, true)
				continue
			}

			text := fmt.Sprintf("🔔 *НОВАЯ ЗАЯВКА (Отдел %d)*\n\n*Услуга:* %s\n*Клиент:* %s", ticket.DepID, ticket.ServiceName, ticket.ClientName)
			callbackData := fmt.Sprintf("accept_%d", ticket.SessionID)
			btn := tgbotapi.NewInlineKeyboardButtonData("✅ Принять заявку", callbackData)

			// Рассылаем заявку только подходящим операторам
			for _, opID := range operators {
				msg := tgbotapi.NewMessage(opID, text)
				msg.ParseMode = "Markdown"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
				c.bot.Send(msg)
			}

			// ИСПРАВЛЕНИЕ: Подтверждаем RabbitMQ успешную обработку заявки
			d.Ack(false)
		}
	}()
}

func (c *Consumer) StartMessageConsumer() {
	ch, err := c.conn.Channel()
	if err != nil {
		return
	}
	q, _ := ch.QueueDeclare("active_chat_messages", true, false, false, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for d := range msgs {
			var msg domain.ChatMessage
			json.Unmarshal(d.Body, &msg)

			clientTgID, operatorTgID, err := c.sessionRepo.GetSessionParticipants(msg.SessionID)
			if err != nil {
				continue
			}

			// Определяем получателя и префикс
			var targetID int64
			var prefix string

			if msg.Sender == "client" && operatorTgID != 0 {
				targetID = operatorTgID
				prefix = "👤 *Клиент:*\n"
			} else if msg.Sender == "operator" {
				targetID = clientTgID
				prefix = "👨‍💻 *Специалист:*\n"
			} else {
				continue // Если получатель неизвестен
			}

			// Безопасное форматирование текста
			caption := prefix + msg.Text
			if msg.Text == "" {
				caption = prefix // Если текста нет (например, просто фото)
			}

			var tgMsg tgbotapi.Chattable

			// Формируем сообщение в зависимости от типа медиа
			switch msg.MediaType {
			case "photo":
				photo := tgbotapi.NewPhoto(targetID, tgbotapi.FileID(msg.FileID))
				photo.Caption = caption
				photo.ParseMode = "Markdown"
				tgMsg = photo
			case "video":
				video := tgbotapi.NewVideo(targetID, tgbotapi.FileID(msg.FileID))
				video.Caption = caption
				video.ParseMode = "Markdown"
				tgMsg = video
			case "voice":
				voice := tgbotapi.NewVoice(targetID, tgbotapi.FileID(msg.FileID))
				voice.Caption = prefix + "*(Голосовое сообщение)*"
				voice.ParseMode = "Markdown"
				tgMsg = voice
			default:
				// Обычный текст
				textMsg := tgbotapi.NewMessage(targetID, caption)
				textMsg.ParseMode = "Markdown"
				tgMsg = textMsg
			}

			c.bot.Send(tgMsg)
		}
	}()
}
