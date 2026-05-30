package delivery

import (
	"log"

	"asakabankbot/internal/domain"
	"asakabankbot/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) handleClientLogic(msg *tgbotapi.Message, tgID int64, text, mediaType, fileID string) {
	user, _ := h.userRepo.GetUserByTelegramID(tgID)
	lang := "ru"
	if user != nil && user.LanguageCode != "" {
		lang = user.LanguageCode
	}

	if msg.IsCommand() && msg.Command() == "start" {
		if user == nil {
			h.userRepo.CreateUser(tgID, msg.From.UserName)
			h.userRepo.UpdateUser(tgID, "", "", "reg_menu")

			welcomeText := i18n.Get("ru", "choose_lang") + "\n\n" + i18n.Get("uz", "choose_lang")
			reply := tgbotapi.NewMessage(msg.Chat.ID, welcomeText)
			reply.ReplyMarkup = SettingsKeyboard()
			h.bot.Send(reply)
			return
		}

		if user.TelNumber != "" {
			h.userRepo.UpdateBotState(tgID, "main_menu")
			reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "back_to_main"))
			reply.ReplyMarkup = MainMenuKeyboard(lang)
			h.bot.Send(reply)
			return
		}

		h.userRepo.UpdateBotState(tgID, "reg_menu")
		reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "ask_name"))
		h.bot.Send(reply)
		return
	}

	if user != nil {
		switch user.BotState {
		case "reg_menu":
			if user.Name == "" {
				h.userRepo.UpdateUser(tgID, text, "", "reg_menu")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "ask_phone"))
				btn := tgbotapi.NewKeyboardButtonContact(i18n.Get(lang, "btn_share_contact"))
				reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(btn))
				h.bot.Send(reply)
			} else if user.TelNumber == "" {
				h.userRepo.UpdateUser(tgID, user.Name, text, "main_menu")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "reg_success"))
				reply.ReplyMarkup = MainMenuKeyboard(lang)
				h.bot.Send(reply)
			}

		case "main_menu":
			switch text {
			case i18n.Get(lang, "btn_about"):
				photoURL := "https://www.asakabank.uz/images/about-bank-tower.png"
				photoMsg := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileURL(photoURL))
				photoMsg.Caption = i18n.Get(lang, "about_text")
				h.bot.Send(photoMsg)
				locMsg := tgbotapi.NewLocation(msg.Chat.ID, 41.302325, 69.274154)
				h.bot.Send(locMsg)
			case i18n.Get(lang, "btn_contacts"):
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "contacts_text"))
				h.bot.Send(reply)
			case i18n.Get(lang, "btn_settings"):
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "choose_lang"))
				reply.ReplyMarkup = SettingsKeyboard()
				h.bot.Send(reply)
			case i18n.Get(lang, "btn_support"):
				deps, _ := h.depRepo.GetAllDepartments()
				h.userRepo.UpdateBotState(tgID, "in_dep")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "support_welcome"))
				reply.ReplyMarkup = DepartmentsReplyKeyboard(deps, lang)
				h.bot.Send(reply)
			default:
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "err_command"))
				h.bot.Send(reply)
			}

		case "in_dep":
			if text == i18n.Get(lang, "btn_back") {
				h.userRepo.UpdateBotState(tgID, "main_menu")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "back_to_main"))
				reply.ReplyMarkup = MainMenuKeyboard(lang)
				h.bot.Send(reply)
				return
			}

			deps, _ := h.depRepo.GetAllDepartments()
			var selectedDep *domain.Department
			for _, d := range deps {
				if i18n.Get(lang, d.Name) == text {
					selectedDep = &d
					break
				}
			}

			if selectedDep == nil {
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "err_command"))
				h.bot.Send(reply)
				return
			}

			if selectedDep.Name == "Общие вопросы" {
				textMsg := i18n.Get(lang, "wait_general")
				reply := tgbotapi.NewMessage(msg.Chat.ID, textMsg)
				reply.ParseMode = "Markdown"
				h.bot.Send(reply)
			} else {
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "ask_service"))
				reply.ReplyMarkup = ServicesKeyboard(selectedDep.ID, selectedDep.Services, lang)
				h.bot.Send(reply)
			}

		case "in_chat":
			if text == i18n.Get(lang, "btn_finish_chat") || text == "❌ Отменить поиск" {
				sessionID, err := h.sessionRepo.GetActiveSessionByClientTgID(tgID)
				if err == nil {
					_, operatorTgID, _ := h.sessionRepo.GetSessionParticipants(sessionID)

					if operatorTgID != 0 {
						isAdmin := h.adminRepo.IsAdmin(operatorTgID)
						h.opRepo.UpdateOperatorStatus(operatorTgID, "online")
						opMsg := tgbotapi.NewMessage(operatorTgID, "❌ Клиент завершил/отменил диалог.")
						opMsg.ReplyMarkup = OperatorMenuKeyboard("online", isAdmin)
						h.bot.Send(opMsg)
					}
					h.sessionRepo.CloseSession(sessionID)
				}

				h.userRepo.UpdateBotState(tgID, "main_menu")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "chat_finished"))
				reply.ReplyMarkup = MainMenuKeyboard(lang)
				h.bot.Send(reply)
				return
			}

			sessionID, err := h.sessionRepo.GetActiveSessionByClientTgID(tgID)
			if err == nil {
				_, operatorTgID, _ := h.sessionRepo.GetSessionParticipants(sessionID)

				if operatorTgID == 0 {
					reply := tgbotapi.NewMessage(msg.Chat.ID, "⏳ *Специалист еще не подключился.*\nПожалуйста, ожидайте или отмените поиск.")
					reply.ParseMode = "Markdown"
					btn := tgbotapi.NewKeyboardButton("❌ Отменить поиск")
					reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(btn))
					h.bot.Send(reply)
					return
				}

				h.sessionRepo.UpdateSessionActivity(sessionID)
				errHistory := h.sessionRepo.AppendToChatHistory(sessionID, "client", text, mediaType, fileID)
				if errHistory != nil {
					log.Printf("ОШИБКА АУДИТА: не удалось записать сообщение клиента %d в историю: %v", tgID, errHistory)
				}
				chatMsg := domain.ChatMessage{SessionID: sessionID, Sender: "client", Text: text, MediaType: mediaType, FileID: fileID}
				errPub := h.publisher.SendChatMessage(chatMsg)
				if errPub != nil {
					log.Printf("RABBITMQ ОШИБКА (Клиент): %v", errPub)
				}
			} else {
				h.userRepo.UpdateBotState(tgID, "main_menu")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "back_to_main"))
				reply.ReplyMarkup = MainMenuKeyboard(lang)
				h.bot.Send(reply)
			}
		}
	}
}
