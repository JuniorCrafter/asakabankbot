package domain

// User описывает модель клиента в нашем приложении
type User struct {
	ID           int
	TelegramID   int64
	Username     string
	Name         string
	TelNumber    string
	BotState     string // reg_menu, main_menu, in_dep, in_chat
	LanguageCode string
}
