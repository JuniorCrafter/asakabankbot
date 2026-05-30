package repository

import (
	"database/sql"
	"log"
)

type OperatorRepository struct {
	db *sql.DB
}

// OperatorProfile структура для хранения развернутой статистики
type OperatorProfile struct {
	Name     string
	Username string
	Stack    string
	DepName  string
	Status   string
}

type OperatorAdminView struct {
	TelegramID int64
	Name       string
	Status     string
}

func NewOperatorRepository(db *sql.DB) *OperatorRepository {
	return &OperatorRepository{db: db}
}

func (r *OperatorRepository) GetAllOperatorIDs() ([]int64, error) {
	var ids []int64
	query := `SELECT telegram_id FROM operators`
	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Ошибка при получении операторов: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *OperatorRepository) IsOperator(tgID int64) bool {
	var id int64
	query := `SELECT telegram_id FROM operators WHERE telegram_id = $1`
	err := r.db.QueryRow(query, tgID).Scan(&id)
	return err == nil
}

// UpdateOperatorStatus меняет статус смены
func (r *OperatorRepository) UpdateOperatorStatus(tgID int64, status string) error {
	query := `UPDATE operators SET status = $1::operator_status_enum WHERE telegram_id = $2`
	res, err := r.db.Exec(query, status, tgID)
	if err != nil {
		log.Printf("БД ОШИБКА: Не удалось сменить статус оператора %d на %s: %v", tgID, status, err)
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		log.Printf("ВНИМАНИЕ: Статус оператора %d на %s не изменен (строка не найдена)", tgID, status)
	}
	return err
}

func (r *OperatorRepository) GetOnlineOperatorsByDepartment(depID int) ([]int64, error) {
	var ids []int64
	query := `SELECT telegram_id FROM operators WHERE department_id = $1 AND status = 'online'::operator_status_enum`

	rows, err := r.db.Query(query, depID)
	if err != nil {
		log.Printf("БД ОШИБКА: Сбой поиска online-операторов: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// GetOperatorProfile исправлен с JOIN на LEFT JOIN пользователей
func (r *OperatorRepository) GetOperatorProfile(tgID int64) (*OperatorProfile, error) {
	profile := &OperatorProfile{}

	query := `
		SELECT COALESCE(u.name, 'Не указано'), COALESCE(u.username, 'Не указано'), 
			   COALESCE(o.stack::text, 'Не указано'), COALESCE(d.name, 'Без отдела'), o.status::text
		FROM operators o
		LEFT JOIN users u ON o.telegram_id = u.telegram_id
		LEFT JOIN departments d ON o.department_id = d.id
		WHERE o.telegram_id = $1`

	err := r.db.QueryRow(query, tgID).Scan(
		&profile.Name,
		&profile.Username,
		&profile.Stack,
		&profile.DepName,
		&profile.Status,
	)
	if err != nil {
		log.Printf("БД ОШИБКА ПРОФИЛЯ (Оператор %d): %v", tgID, err)
		return nil, err
	}
	return profile, nil
}
func (r *OperatorRepository) GetOperatorStatusByID(tgID int64) string {
	var status string
	query := `SELECT status::text FROM operators WHERE telegram_id = $1`
	err := r.db.QueryRow(query, tgID).Scan(&status)
	if err != nil {
		log.Printf("БД ОШИБКА (Чтение статуса %d): %v", tgID, err)
		return "offline" // Безопасный фоллбэк
	}
	return status
}

// GetOperatorsForAdmin выгружает список всех сотрудников для панели управления
func (r *OperatorRepository) GetOperatorsForAdmin() ([]OperatorAdminView, error) {
	var list []OperatorAdminView
	query := `
		SELECT o.telegram_id, COALESCE(u.name, 'Не указано'), o.status::text 
		FROM operators o
		LEFT JOIN users u ON o.telegram_id = u.telegram_id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v OperatorAdminView
		if err := rows.Scan(&v.TelegramID, &v.Name, &v.Status); err == nil {
			list = append(list, v)
		}
	}
	return list, nil
}
