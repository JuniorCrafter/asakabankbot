package delivery

import (
	"fmt"

	"asakabankbot/internal/domain"
	"asakabankbot/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainMenuKeyboard возвращает главное меню
func MainMenuKeyboard(lang string) tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		// 1 ряд: Большая кнопка поддержки
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(lang, "btn_support")),
			// 2 ряд: Две кнопки рядом
			tgbotapi.NewKeyboardButton(i18n.Get(lang, "btn_about")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(lang, "btn_settings")),
			// 3 ряд: Настройки
			tgbotapi.NewKeyboardButton(i18n.Get(lang, "btn_contacts")),
		),
	)
	keyboard.ResizeKeyboard = true // Обязательный параметр для компактности
	return keyboard
}

// DepartmentsReplyKeyboard переводит названия отделов
func DepartmentsReplyKeyboard(departments []domain.Department, lang string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for _, dep := range departments {
		// Берем оригинальное русское название из БД и переводим его
		translatedName := i18n.Get(lang, dep.Name)
		btn := tgbotapi.NewKeyboardButton(translatedName)
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(btn))
	}

	backBtn := tgbotapi.NewKeyboardButton(i18n.Get(lang, "btn_back"))
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(backBtn))

	return tgbotapi.ReplyKeyboardMarkup{Keyboard: rows, ResizeKeyboard: true}
}

// ServicesKeyboard переводит названия услуг внутри инлайн-кнопок
func ServicesKeyboard(depID int, services []string, lang string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i, serviceName := range services {
		callbackData := fmt.Sprintf("srv_%d_%d", depID, i)
		translatedService := i18n.Get(lang, serviceName)
		btn := tgbotapi.NewInlineKeyboardButtonData(translatedService, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// SettingsKeyboard остается статичной, так как флаги и названия языков универсальны
func SettingsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇺🇿 O'zbekcha", "lang_uz"),
		),
	)
}

// OperatorMenuKeyboard динамически формирует кнопки
func OperatorMenuKeyboard(status string, isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	var statusBtn tgbotapi.KeyboardButton

	if status == "online" {
		statusBtn = tgbotapi.NewKeyboardButton("🔴 Завершить смену (Offline)")
	} else {
		statusBtn = tgbotapi.NewKeyboardButton("🟢 Начать смену (Online)")
	}

	rows := [][]tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButtonRow(statusBtn),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("📊 Моя статистика")),
	}

	// Если есть права админа, добавляем коронную кнопку
	if isAdmin {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("👑 Панель Администратора")))
	}

	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// AdminMenuKeyboard возвращает обновленное меню панели управления
func AdminMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить оператора"),
			tgbotapi.NewKeyboardButton("❌ Удалить оператора"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔄 Изменить статус оператора"),
			tgbotapi.NewKeyboardButton("📈 Глобальная статистика"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Вернуться к смене"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}
