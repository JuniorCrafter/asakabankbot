package domain

// ChatTicket — это заявка, которая отправляется операторам через RabbitMQ
type ChatTicket struct {
	SessionID   int    `json:"session_id"`
	ClientTgID  int64  `json:"client_tg_id"`
	ClientName  string `json:"client_name"`
	DepID       int    `json:"dep_id"`
	ServiceName string `json:"service_name"`
}
