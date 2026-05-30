package delivery

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"asakabankbot/internal/broker"
	"asakabankbot/internal/domain"
	"asakabankbot/internal/i18n"
	"asakabankbot/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot         *tgbotapi.BotAPI
	userRepo    *repository.UserRepository
	depRepo     *repository.DepartmentRepository
	sessionRepo *repository.SessionRepository
	opRepo      *repository.OperatorRepository
	adminRepo   *repository.AdminRepository // Добавлено хранилище администратора
	publisher   *broker.Publisher
	limiter     *RateLimiter
}

func NewBotHandler(bot *tgbotapi.BotAPI, userRepo *repository.UserRepository, depRepo *repository.DepartmentRepository, sessionRepo *repository.SessionRepository, opRepo *repository.OperatorRepository, adminRepo *repository.AdminRepository, publisher *broker.Publisher) *BotHandler {
	return &BotHandler{
		bot:         bot,
		userRepo:    userRepo,
		depRepo:     depRepo,
		sessionRepo: sessionRepo,
		opRepo:      opRepo,
		adminRepo:   adminRepo,
		publisher:   publisher,
		limiter:     NewRateLimiter(10, 1*time.Minute), // 10 запросов в минуту
	}
}

func StartBot(token string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка при инициализации бота: %v", err)
	}
	bot.Debug = false
	return bot
}

func (h *BotHandler) HandleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		go h.processUpdate(update)
	}
}

func (h *BotHandler) processUpdate(update tgbotapi.Update) {

	var tgID int64

	// 1. Безопасное извлечение ID
	if update.Message != nil {
		tgID = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		tgID = update.CallbackQuery.From.ID
	} else {
		return // Игнорируем другие типы апдейтов
	}

	if !h.limiter.Allow(tgID) {
		log.Printf("БЛОКИРОВКА: Пользователь %d превысил лимит запросов (Спам)", tgID)
		return
	}

	// 2. Обработка кнопок
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// 3. Защита от пустых сообщений (на всякий случай)
	if update.Message == nil {
		return
	}

	// 4. ТЕПЕРЬ БЕЗОПАСНО ИЗВЛЕКАТЬ MSG
	msg := update.Message

	// --- ИСПРАВЛЕНИЕ: Распознавание типа сообщения и извлечение FileID ---
	var text, mediaType, fileID string
	mediaType = "text"

	if msg.Text != "" {
		text = msg.Text
	} else if len(msg.Photo) > 0 {
		mediaType = "photo"
		fileID = msg.Photo[len(msg.Photo)-1].FileID
		text = msg.Caption
	} else if msg.Video != nil {
		mediaType = "video"
		fileID = msg.Video.FileID
		text = msg.Caption
	} else if msg.Voice != nil {
		mediaType = "voice"
		fileID = msg.Voice.FileID
	} else if msg.Contact != nil {
		text = msg.Contact.PhoneNumber
	} else if !msg.IsCommand() {
		h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Этот формат не поддерживается. Разрешены только текст, фото, видео и голосовые сообщения."))
		return
	}

	if text == "" && mediaType == "text" && !msg.IsCommand() {
		return
	}
	// ==========================================
	// 1. ИЗОЛИРОВАННАЯ ЛОГИКА ОПЕРАТОРА / АДМИНА
	// ==========================================

	if h.opRepo.IsOperator(tgID) {
		h.handleOperatorLogic(msg, tgID, text, mediaType, fileID)
		return
	}

	h.handleClientLogic(msg, tgID, text, mediaType, fileID)
}

func (h *BotHandler) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(query.ID, "")
	h.bot.Request(callback)

	tgID := query.From.ID
	data := query.Data

	user, err := h.userRepo.GetUserByTelegramID(tgID)
	if err != nil || user == nil {
		return
	}

	langCode := "ru"
	if user.LanguageCode != "" {
		langCode = user.LanguageCode
	}

	if data == "lang_ru" || data == "lang_uz" {
		if data == "lang_uz" {
			langCode = "uz"
		} else {
			langCode = "ru"
		}

		h.userRepo.UpdateLanguage(tgID, langCode)

		if user.Name == "" {
			h.userRepo.UpdateBotState(tgID, "reg_menu")
			h.bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))

			msg := tgbotapi.NewMessage(query.Message.Chat.ID, i18n.Get(langCode, "ask_name"))
			h.bot.Send(msg)
			return
		}

		h.bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, i18n.Get(langCode, "lang_changed"))
		msg.ReplyMarkup = MainMenuKeyboard(langCode)
		h.bot.Send(msg)
		return
	}

	// --- ОБРАБОТКА ПРИНУДИТЕЛЬНОГО СБРОСА СТАТУСА АДМИНИСТРАТОРОМ ---
	if strings.HasPrefix(data, "setstat_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			status := parts[1] // offline
			opTgID, _ := strconv.ParseInt(parts[2], 10, 64)

			err := h.opRepo.UpdateOperatorStatus(opTgID, status)
			if err == nil {
				h.bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))

				successMsg := tgbotapi.NewMessage(query.Message.Chat.ID, fmt.Sprintf("✅ Статус оператора `%d` изменен на `%s`", opTgID, status))
				h.bot.Send(successMsg)

				// ИСПРАВЛЕНИЕ: Если админ сбросил статус самому себе, обновляем клавиатуру
				if opTgID == tgID {
					updateMsg := tgbotapi.NewMessage(query.Message.Chat.ID, "Вы вышли из панели управления. Клавиатура обновлена.")
					updateMsg.ReplyMarkup = OperatorMenuKeyboard(status, true)
					h.bot.Send(updateMsg)
				}
			} else {
				h.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "❌ Ошибка БД."))
			}
		}
		return
	}

	if strings.HasPrefix(data, "accept_") {
		parts := strings.Split(data, "_")
		if len(parts) == 2 {
			sessionID, err := strconv.Atoi(parts[1])
			if err == nil {
				clientTgID, err := h.sessionRepo.AssignOperator(sessionID, tgID)
				if err == nil {
					h.opRepo.UpdateOperatorStatus(tgID, "busy")

					h.bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))

					opMsgText := fmt.Sprintf("✅ Вы успешно подключились к сессии #%d.\nТеперь все ваши сообщения будут отправляться клиенту.", sessionID)
					opMsg := tgbotapi.NewMessage(query.Message.Chat.ID, opMsgText)
					opMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🏁 Завершить диалог")),
					)
					h.bot.Send(opMsg)

					clientMsg := tgbotapi.NewMessage(clientTgID, "🟢 Специалист подключился к диалогу. Вы можете задать свой вопрос.")
					h.bot.Send(clientMsg)
				} else {
					h.bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))
					errorMsg := tgbotapi.NewMessage(query.Message.Chat.ID, "❌ Заявка уже принята другим оператором или завершена.")
					h.bot.Send(errorMsg)
				}
			}
		}
		return
	}

	if strings.HasPrefix(data, "srv_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			depID, _ := strconv.Atoi(parts[1])
			serviceIndex, _ := strconv.Atoi(parts[2])

			dep, err := h.depRepo.GetDepartmentByID(depID)
			if err == nil && serviceIndex < len(dep.Services) {
				serviceName := dep.Services[serviceIndex]

				sessionID, err := h.sessionRepo.CreateSession(tgID, depID, serviceName)
				if err != nil {
					return
				}

				ticket := domain.ChatTicket{
					SessionID:   sessionID,
					ClientTgID:  tgID,
					ClientName:  user.Name,
					DepID:       depID,
					ServiceName: serviceName,
				}
				h.publisher.SendTicket(ticket)
				h.userRepo.UpdateBotState(tgID, "in_chat")

				h.bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))

				translatedService := i18n.Get(langCode, serviceName)
				text := fmt.Sprintf(i18n.Get(langCode, "wait_operator"), translatedService)

				newMsg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
				newMsg.ParseMode = "Markdown"
				newMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(i18n.Get(langCode, "btn_finish_chat"))),
				)
				h.bot.Send(newMsg)
			}
		}
	}
}
