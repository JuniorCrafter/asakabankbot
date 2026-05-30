package repository

import (
	"database/sql"
	"fmt"
	"log"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// IsAdmin проверяет наличие прав администратора
func (r *AdminRepository) IsAdmin(tgID int64) bool {
	var isAdmin bool
	query := `SELECT is_admin FROM operators WHERE telegram_id = $1`
	err := r.db.QueryRow(query, tgID).Scan(&isAdmin)
	if err != nil {
		return false
	}
	return isAdmin
}

// GetGlobalStats собирает статистику со всех таблиц
func (r *AdminRepository) GetGlobalStats() string {
	var totalUsers, activeChats, onlineOps int

	_ = r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM chat_sessions WHERE status IN ('active'::session_status_enum, 'in_progress'::session_status_enum)`).Scan(&activeChats)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM operators WHERE status = 'online'::operator_status_enum`).Scan(&onlineOps)

	return fmt.Sprintf("📈 *ГЛОБАЛЬНАЯ СТАТИСТИКА*\n\n👥 Зарегистрировано клиентов: %d\n💬 Открытых диалогов: %d\n🟢 Операторов на смене: %d", totalUsers, activeChats, onlineOps)
}

// AddOperator регистрирует нового оператора
func (r *AdminRepository) AddOperator(tgID int64, depID int, stack string) error {
	query := `
		INSERT INTO operators (telegram_id, stack, status, department_id, is_admin) 
		VALUES ($1, $2::operator_stack_enum, 'offline'::operator_status_enum, $3, false)
		ON CONFLICT (telegram_id) 
		DO UPDATE SET stack = $2::operator_stack_enum, department_id = $3`

	_, err := r.db.Exec(query, tgID, stack, depID)
	if err != nil {
		log.Printf("БД ОШИБКА (Добавление оператора): %v", err)
	}
	return err
}

// RemoveOperator удаляет права оператора
func (r *AdminRepository) RemoveOperator(tgID int64) error {
	query := `DELETE FROM operators WHERE telegram_id = $1`
	_, err := r.db.Exec(query, tgID)
	if err != nil {
		log.Printf("БД ОШИБКА (Удаление оператора): %v", err)
	}
	return err
}
