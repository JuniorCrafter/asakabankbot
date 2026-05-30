package delivery

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) handleAdminLogic(msg *tgbotapi.Message, tgID int64, text string) {
	isAdmin := h.adminRepo.IsAdmin(tgID)

	switch text {
	case "🔙 Вернуться к смене":
		h.opRepo.UpdateOperatorStatus(tgID, "offline")
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Вы вернулись в меню оператора.")
		reply.ReplyMarkup = OperatorMenuKeyboard("offline", isAdmin)
		h.bot.Send(reply)

	case "📈 Глобальная статистика":
		stats := h.adminRepo.GetGlobalStats()
		reply := tgbotapi.NewMessage(msg.Chat.ID, stats)
		reply.ParseMode = "Markdown"
		h.bot.Send(reply)

	case "➕ Добавить оператора":
		helpText := "➕ *Добавление нового оператора*\n\n" +
			"Отправьте текстовую команду в чат строго в формате:\n" +
			"`/addop <Telegram_ID> <ID_Отдела> <Стек>`\n\n" +
			"*ID отделов:* 1 (Физ. лица), 2 (Юр. лица), 3 (Махалла банкирлари), 4 (Общие вопросы)\n" +
			"*Доступные стеки:* `expert`, `expert_pro`, `expert_vip`, `expert_lite`"
		reply := tgbotapi.NewMessage(msg.Chat.ID, helpText)
		reply.ParseMode = "Markdown"
		h.bot.Send(reply)

	case "❌ Удалить оператора":
		helpText := "❌ *Удаление оператора из системы*\n\n" +
			"Отправьте текстовую команду в чат строго в формате:\n" +
			"`/delop <Telegram_ID>`"
		reply := tgbotapi.NewMessage(msg.Chat.ID, helpText)
		reply.ParseMode = "Markdown"
		h.bot.Send(reply)

	case "🔄 Изменить статус оператора":
		ops, err := h.opRepo.GetOperatorsForAdmin()
		if err != nil || len(ops) == 0 {
			h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Список операторов пуст или произошла ошибка."))
			return
		}

		h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "📋 *Список сотрудников в системе:*\nНажмите кнопку под нужным оператором для принудительного перевода в Offline."))
		for _, op := range ops {
			cardText := fmt.Sprintf("👤 *%s*\n🆔 ID: `%d`\n📊 Статус: `%s`", op.Name, op.TelegramID, op.Status)
			inlineBtn := tgbotapi.NewInlineKeyboardButtonData("🛑 Сбросить в Offline", fmt.Sprintf("setstat_offline_%d", op.TelegramID))
			inlineMarkup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(inlineBtn))

			reply := tgbotapi.NewMessage(msg.Chat.ID, cardText)
			reply.ParseMode = "Markdown"
			reply.ReplyMarkup = inlineMarkup
			h.bot.Send(reply)
		}

	default:
		if strings.HasPrefix(text, "/addop") {
			parts := strings.Split(text, " ")
			if len(parts) == 4 {
				opID, _ := strconv.ParseInt(parts[1], 10, 64)
				depID, _ := strconv.Atoi(parts[2])
				stack := parts[3]

				err := h.adminRepo.AddOperator(opID, depID, stack)
				if err != nil {
					h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка. Проверьте параметры."))
				} else {
					h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "✅ Оператор успешно добавлен/обновлен!"))
				}
			}
		} else if strings.HasPrefix(text, "/delop") {
			parts := strings.Split(text, " ")
			if len(parts) == 2 {
				opID, _ := strconv.ParseInt(parts[1], 10, 64)
				err := h.adminRepo.RemoveOperator(opID)
				if err != nil {
					h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка при удалении."))
				} else {
					h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "✅ Оператор удален!"))
				}
			}
		}
	}
}
