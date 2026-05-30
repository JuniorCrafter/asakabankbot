package delivery

import (
	"fmt"
	"log"
	"strings"

	"asakabankbot/internal/domain"
	"asakabankbot/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) handleOperatorLogic(msg *tgbotapi.Message, tgID int64, text, mediaType, fileID string) {
	currentStatus := h.opRepo.GetOperatorStatusByID(tgID)
	isAdmin := h.adminRepo.IsAdmin(tgID)

	if msg.IsCommand() && msg.Command() == "start" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "👨‍💻 Рабочее место оператора")
		reply.ReplyMarkup = OperatorMenuKeyboard(currentStatus, isAdmin)
		h.bot.Send(reply)
		return
	}

	if currentStatus == "admin_menu" {
		h.handleAdminLogic(msg, tgID, text)
		return
	}

	sessionID, clientTgID, err := h.sessionRepo.GetActiveSessionByOperator(tgID)
	if err == nil && sessionID != 0 {
		if text == "🏁 Завершить диалог" {
			h.sessionRepo.CloseSession(sessionID)

			clientUser, _ := h.userRepo.GetUserByTelegramID(clientTgID)
			clientLang := "ru"
			if clientUser != nil && clientUser.LanguageCode != "" {
				clientLang = clientUser.LanguageCode
			}
			h.userRepo.UpdateBotState(clientTgID, "main_menu")
			clientMsg := tgbotapi.NewMessage(clientTgID, i18n.Get(clientLang, "chat_finished"))
			clientMsg.ReplyMarkup = MainMenuKeyboard(clientLang)
			h.bot.Send(clientMsg)

			h.opRepo.UpdateOperatorStatus(tgID, "online")
			reply := tgbotapi.NewMessage(msg.Chat.ID, "✅ Диалог успешно завершен. Вы вернулись на смену.")
			reply.ReplyMarkup = OperatorMenuKeyboard("online", isAdmin)
			h.bot.Send(reply)
			return
		}

		h.sessionRepo.UpdateSessionActivity(sessionID)
		errHistory := h.sessionRepo.AppendToChatHistory(sessionID, "operator", text, mediaType, fileID)
		if errHistory != nil {
			log.Printf("ОШИБКА АУДИТА: не удалось записать сообщение оператора %d в историю: %v", tgID, errHistory)
		}
		chatMsg := domain.ChatMessage{SessionID: sessionID, Sender: "operator", Text: text, MediaType: mediaType, FileID: fileID}
		errPub := h.publisher.SendChatMessage(chatMsg)
		if errPub != nil {
			log.Printf("RABBITMQ ОШИБКА (Оператор): %v", errPub)
		}
		return
	}

	switch text {
	case "🟢 Начать смену (Online)":
		h.opRepo.UpdateOperatorStatus(tgID, "online")
		reply := tgbotapi.NewMessage(msg.Chat.ID, "✅ Вы вышли на смену.")
		reply.ReplyMarkup = OperatorMenuKeyboard("online", isAdmin)
		h.bot.Send(reply)
	case "🔴 Завершить смену (Offline)":
		h.opRepo.UpdateOperatorStatus(tgID, "offline")
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⏸ Смена завершена. Поступление заявок приостановлено.")
		reply.ReplyMarkup = OperatorMenuKeyboard("offline", isAdmin)
		h.bot.Send(reply)
	case "📊 Моя статистика":
		prof, err := h.opRepo.GetOperatorProfile(tgID)
		if err == nil {
			safeStack := strings.ReplaceAll(prof.Stack, "_", "\\_")
			safeUsername := strings.ReplaceAll(prof.Username, "_", "\\_")

			statsText := fmt.Sprintf(
				"📋 *ПРОФИЛЬ СОТРУДНИКА*\n\n*ФИО:* %s\n*Логин:* @%s\n*Должность:* %s\n*Департамент:* %s\n*Текущий статус:* %s",
				prof.Name, safeUsername, safeStack, prof.DepName, prof.Status,
			)
			reply := tgbotapi.NewMessage(msg.Chat.ID, statsText)
			reply.ParseMode = "Markdown"
			reply.ReplyMarkup = OperatorMenuKeyboard(currentStatus, isAdmin)
			h.bot.Send(reply)
		} else {
			reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Не удалось загрузить статистику.")
			h.bot.Send(reply)
		}
	case "👑 Панель Администратора":
		if isAdmin {
			h.opRepo.UpdateOperatorStatus(tgID, "admin_menu")
			reply := tgbotapi.NewMessage(msg.Chat.ID, "👑 Добро пожаловать в панель управления.")
			reply.ReplyMarkup = AdminMenuKeyboard()
			h.bot.Send(reply)
		}
	}
}
