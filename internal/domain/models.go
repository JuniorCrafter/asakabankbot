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

// Department описывает структуру банковского отдела
type Department struct {
	ID       int
	Name     string
	Services []string // Сюда мы будем распаковывать JSON из базы данных
}

type ChatMessage struct {
	SessionID int    `json:"session_id"`
	Sender    string `json:"sender"`     // "client" или "operator"
	Text      string `json:"text"`       // Для текста или подписи (caption) к фото/видео
	MediaType string `json:"media_type"` // "text", "photo", "video", "voice"
	FileID    string `json:"file_id"`    // Уникальный ID файла на серверах Telegram
}

// ChatTicket — это заявка, которая отправляется операторам через RabbitMQ
type ChatTicket struct {
	SessionID   int    `json:"session_id"`
	ClientTgID  int64  `json:"client_tg_id"`
	ClientName  string `json:"client_name"`
	DepID       int    `json:"dep_id"`
	ServiceName string `json:"service_name"`
}
