package repository

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// ExpiredSession структура для уведомления участников при автозакрытии
type ExpiredSession struct {
	SessionID    int
	ClientTgID   int64
	OperatorID   sql.NullInt64
	OperatorID64 int64
}

// CreateSession создает сессию и возвращает её ID
func (r *SessionRepository) CreateSession(clientTgID int64, depID int, serviceName string) (int, error) {
	var sessionID int
	query := `INSERT INTO chat_sessions (client_telegram_id, department_id, service, status, chat_history) 
              VALUES ($1, $2, $3, 'active'::session_status_enum, '[]'::jsonb) RETURNING id`

	err := r.db.QueryRow(query, clientTgID, depID, serviceName).Scan(&sessionID)
	if err != nil {
		log.Printf("Ошибка при создании сессии чата: %v", err)
		return 0, err
	}
	return sessionID, nil
}

// GetActiveSessionByClientTgID ищет ID активной сессии для клиента
func (r *SessionRepository) GetActiveSessionByClientTgID(clientTgID int64) (int, error) {
	var sessionID int
	query := `SELECT id FROM chat_sessions 
              WHERE client_telegram_id = $1 AND status IN ('active'::session_status_enum, 'in_progress'::session_status_enum) 
              ORDER BY id DESC LIMIT 1`

	err := r.db.QueryRow(query, clientTgID).Scan(&sessionID)
	if err != nil {
		return 0, err
	}
	return sessionID, nil
}

// AssignOperator закрепляет оператора за сессией и возвращает Telegram ID клиента
func (r *SessionRepository) AssignOperator(sessionID int, operatorTgID int64) (int64, error) {
	var clientTgID int64

	// Модифицировано: находим внутренний id оператора по его telegram_id через подзапрос
	query := `UPDATE chat_sessions 
              SET operator_id = (SELECT id FROM operators WHERE telegram_id = $1), 
                  status = 'in_progress'::session_status_enum 
              WHERE id = $2 AND status = 'active'::session_status_enum 
              RETURNING client_telegram_id`

	err := r.db.QueryRow(query, operatorTgID, sessionID).Scan(&clientTgID)
	if err != nil {
		log.Printf("Ошибка при назначении оператора на сессию %d: %v", sessionID, err)
		return 0, err
	}

	return clientTgID, nil
}

// GetSessionParticipants возвращает ID клиента и ID оператора
func (r *SessionRepository) GetSessionParticipants(sessionID int) (int64, int64, error) {
	var clientTgID int64
	var opTgID sql.NullInt64

	// Делаем JOIN с таблицей operators, чтобы достать telegram_id по внутреннему operator_id
	query := `SELECT cs.client_telegram_id, o.telegram_id 
              FROM chat_sessions cs
              LEFT JOIN operators o ON cs.operator_id = o.id 
              WHERE cs.id = $1`

	err := r.db.QueryRow(query, sessionID).Scan(&clientTgID, &opTgID)
	if err != nil {
		log.Printf("БД ОШИБКА (GetSessionParticipants): %v", err)
	}
	return clientTgID, opTgID.Int64, err
}

// GetActiveSessionByOperator ищет активную сессию, которую сейчас ведет оператор
func (r *SessionRepository) GetActiveSessionByOperator(operatorTgID int64) (int, int64, error) {
	var sessionID int
	var clientTgID int64

	// Модифицировано: выборка идет через JOIN по telegram_id оператора
	query := `SELECT cs.id, cs.client_telegram_id 
              FROM chat_sessions cs 
              JOIN operators o ON cs.operator_id = o.id
              WHERE o.telegram_id = $1 AND cs.status = 'in_progress'::session_status_enum 
              ORDER BY cs.id DESC LIMIT 1`

	err := r.db.QueryRow(query, operatorTgID).Scan(&sessionID, &clientTgID)
	if err != nil {
		return 0, 0, err
	}
	return sessionID, clientTgID, err
}

func (r *SessionRepository) CloseSession(sessionID int) error {
	// Модифицировано: добавлено приведение типов ENUM и фиксация времени закрытия closed_at
	query := `UPDATE chat_sessions 
              SET status = 'closed'::session_status_enum, 
                  closed_at = CURRENT_TIMESTAMP 
              WHERE id = $1`

	res, err := r.db.Exec(query, sessionID)
	if err != nil {
		log.Printf("DB ОШИБКА (Закрытие сессии %d): %v", sessionID, err)
		return err
	}

	rows, _ := res.RowsAffected()
	log.Printf("БД ОТЧЕТ: Сессия %d переведена в status='closed'. Изменено строк: %d", sessionID, rows)
	return nil
}

// UpdateSessionActivity обновляет временную метку последней активности в чате
func (r *SessionRepository) UpdateSessionActivity(sessionID int) {
	query := `UPDATE chat_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, _ = r.db.Exec(query, sessionID)
}

// CloseExpiredSessions автоматически закрывает сессии без активности и возвращает данные участников
func (r *SessionRepository) CloseExpiredSessions(minutes int) ([]ExpiredSession, error) {
	// Закрываем сессии и возвращаем ID участников, чтобы бот мог отправить им уведомления
	query := `
		UPDATE chat_sessions 
		SET status = 'closed'::session_status_enum, 
		    closed_at = CURRENT_TIMESTAMP 
		WHERE status IN ('active'::session_status_enum, 'in_progress'::session_status_enum) 
		  AND updated_at < NOW() - ($1 || ' minutes')::interval
		RETURNING id, client_telegram_id, operator_id`

	rows, err := r.db.Query(query, minutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []ExpiredSession
	for rows.Next() {
		var s ExpiredSession
		if err := rows.Scan(&s.SessionID, &s.ClientTgID, &s.OperatorID); err == nil {
			// Если оператор был назначен, подтягиваем его telegram_id
			if s.OperatorID.Valid {
				_ = r.db.QueryRow(`SELECT telegram_id FROM operators WHERE id = $1`, s.OperatorID.Int64).Scan(&s.OperatorID64)
			}
			expired = append(expired, s)
		}
	}
	return expired, nil
}

// IsSessionActive проверяет, не отменил ли клиент заявку
func (r *SessionRepository) IsSessionActive(sessionID int) bool {
	var status string
	query := `SELECT status::text FROM chat_sessions WHERE id = $1`
	err := r.db.QueryRow(query, sessionID).Scan(&status)
	if err != nil {
		return false
	}
	return status == "active" || status == "in_progress"
}

// AppendToChatHistory добавляет одно сообщение в JSONB массив истории сессии
func (r *SessionRepository) AppendToChatHistory(sessionID int, sender, text, mediaType, fileID string) error {
	// 1. Формируем структуру одного сообщения
	historyMsg := map[string]interface{}{
		"sender":     sender,
		"text":       text,
		"media_type": mediaType,
		"file_id":    fileID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// 2. Оборачиваем в массив из одного элемента (требование PostgreSQL для конкатенации массивов)
	jsonBytes, err := json.Marshal([]interface{}{historyMsg})
	if err != nil {
		log.Printf("Ошибка сериализации истории: %v", err)
		return err
	}

	// 3. Оператор || склеивает старый JSONB-массив с новым
	query := `UPDATE chat_sessions SET chat_history = chat_history || $1::jsonb WHERE id = $2`
	_, err = r.db.Exec(query, string(jsonBytes), sessionID)
	if err != nil {
		log.Printf("БД ОШИБКА (Сохранение истории сессии %d): %v", sessionID, err)
	}
	return err
}
