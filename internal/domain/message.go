package domain

type ChatMessage struct {
	SessionID int    `json:"session_id"`
	Sender    string `json:"sender"`     // "client" или "operator"
	Text      string `json:"text"`       // Для текста или подписи (caption) к фото/видео
	MediaType string `json:"media_type"` // "text", "photo", "video", "voice"
	FileID    string `json:"file_id"`    // Уникальный ID файла на серверах Telegram
}
