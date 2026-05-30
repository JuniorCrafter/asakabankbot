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
		var tgID int64

		// 1. Безопасное извлечение ID
		if update.Message != nil {
			tgID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			tgID = update.CallbackQuery.From.ID
		} else {
			continue // Игнорируем другие типы апдейтов
		}

		if !h.limiter.Allow(tgID) {
			log.Printf("БЛОКИРОВКА: Пользователь %d превысил лимит запросов (Спам)", tgID)
			continue
		}

		// 2. Обработка кнопок
		if update.CallbackQuery != nil {
			h.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		// 3. Защита от пустых сообщений (на всякий случай)
		if update.Message == nil {
			continue
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
			continue
		}

		if text == "" && mediaType == "text" && !msg.IsCommand() {
			continue
		}
		// ==========================================
		// 1. ИЗОЛИРОВАННАЯ ЛОГИКА ОПЕРАТОРА / АДМИНА
		// ==========================================
		if h.opRepo.IsOperator(tgID) {
			currentStatus := h.opRepo.GetOperatorStatusByID(tgID)
			isAdmin := h.adminRepo.IsAdmin(tgID)

			if msg.IsCommand() && msg.Command() == "start" {
				reply := tgbotapi.NewMessage(msg.Chat.ID, "👨‍💻 Рабочее место оператора")
				reply.ReplyMarkup = OperatorMenuKeyboard(currentStatus, isAdmin)
				h.bot.Send(reply)
				continue
			}

			// --- РЕЖИМ АДМИНИСТРАТОРА ---
			if currentStatus == "admin_menu" {
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
						continue
					}

					h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "📋 *Список сотрудников в системе:*\nНажмите кнопку под нужным оператором для принудительного перевода в Offline."))
					for _, op := range ops {
						cardText := fmt.Sprintf("👤 *%s*\n🆔 ID: `%d`\n📊 Статус: `%s`", op.Name, op.TelegramID, op.Status)

						// Кнопка принудительного сброса статуса для каждого оператора индивидуально
						inlineBtn := tgbotapi.NewInlineKeyboardButtonData("🛑 Сбросить в Offline", fmt.Sprintf("setstat_offline_%d", op.TelegramID))
						inlineMarkup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(inlineBtn))

						reply := tgbotapi.NewMessage(msg.Chat.ID, cardText)
						reply.ParseMode = "Markdown"
						reply.ReplyMarkup = inlineMarkup
						h.bot.Send(reply)
					}

				default:
					// Обработка текстовых команд добавления/удаления (оставляем без изменений)
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
				continue
			}

			// --- РЕЖИМ ОПЕРАТОРА В ЧАТЕ ---
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
					continue
				}

				h.sessionRepo.UpdateSessionActivity(sessionID)
				h.sessionRepo.AppendToChatHistory(sessionID, "operator", text, mediaType, fileID)
				errHistory := h.sessionRepo.AppendToChatHistory(sessionID, "operator", text, mediaType, fileID)
				if errHistory != nil {
					log.Printf("ОШИБКА АУДИТА: не удалось записать сообщение оператора %d в историю: %v", tgID, errHistory)
				}
				chatMsg := domain.ChatMessage{SessionID: sessionID, Sender: "operator", Text: text, MediaType: mediaType, FileID: fileID}
				errPub := h.publisher.SendChatMessage(chatMsg)
				if errPub != nil {
					log.Printf("RABBITMQ ОШИБКА (Оператор): %v", errPub)
				}
				continue
			}

			// --- РЕЖИМ МЕНЮ ОПЕРАТОРА ---
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
					// Делаем текст безопасным для Telegram Markdown, экранируя подчеркивания
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
			continue
		}

		// ==========================================
		// 2. ЛОГИКА КЛИЕНТА
		// ==========================================
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
				continue
			}

			if user.TelNumber != "" {
				h.userRepo.UpdateBotState(tgID, "main_menu")
				reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "back_to_main"))
				reply.ReplyMarkup = MainMenuKeyboard(lang)
				h.bot.Send(reply)
				continue
			}

			h.userRepo.UpdateBotState(tgID, "reg_menu")
			reply := tgbotapi.NewMessage(msg.Chat.ID, i18n.Get(lang, "ask_name"))
			h.bot.Send(reply)
			continue
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
					continue
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
					continue
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
					continue
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
						continue
					}

					h.sessionRepo.UpdateSessionActivity(sessionID)
					h.sessionRepo.AppendToChatHistory(sessionID, "client", text, mediaType, fileID)
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
