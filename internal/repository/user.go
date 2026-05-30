package repository

import (
	"database/sql"
	"errors"
	"log"

	"asakabankbot/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByTelegramID(tgID int64) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, telegram_id, COALESCE(username, ''), COALESCE(name, ''), 
              COALESCE(tel_number, ''), bot_state::text, COALESCE(language_code, '') 
              FROM users WHERE telegram_id = $1`

	err := r.db.QueryRow(query, tgID).Scan(
		&user.ID, &user.TelegramID, &user.Username,
		&user.Name, &user.TelNumber, &user.BotState, &user.LanguageCode,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) CreateUser(tgID int64, username string) error {
	query := `INSERT INTO users (telegram_id, username, bot_state, language_code) 
              VALUES ($1, $2, 'reg_menu'::bot_state_enum, 'ru') 
              ON CONFLICT (telegram_id) DO NOTHING`

	_, err := r.db.Exec(query, tgID, username)
	if err != nil {
		log.Printf("БД ОШИБКА (Создание пользователя %d): %v", tgID, err)
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(tgID int64, name, telNumber, botState string) error {
	query := `UPDATE users SET name = $1, tel_number = $2, bot_state = $3::bot_state_enum WHERE telegram_id = $4`
	res, err := r.db.Exec(query, name, telNumber, botState, tgID)
	if err != nil {
		log.Printf("БД ОШИБКА (Обновление пользователя %d): %v", tgID, err)
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		log.Printf("ВНИМАНИЕ: Пользователь %d не обновлен (возможно, не найден)", tgID)
	}
	return nil
}

func (r *UserRepository) UpdateLanguage(tgID int64, langCode string) error {
	query := `UPDATE users SET language_code = $1 WHERE telegram_id = $2`
	_, err := r.db.Exec(query, langCode, tgID)
	if err != nil {
		log.Printf("БД ОШИБКА (Обновление языка %d): %v", tgID, err)
		return err
	}
	return nil
}

func (r *UserRepository) UpdateBotState(tgID int64, state string) error {
	query := `UPDATE users SET bot_state = $1::bot_state_enum WHERE telegram_id = $2`
	res, err := r.db.Exec(query, state, tgID)
	if err != nil {
		log.Printf("БД ОШИБКА (Смена статуса пользователя %d на %s): %v", tgID, state, err)
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		log.Printf("ВНИМАНИЕ: Статус пользователя %d не изменен (запись не найдена)", tgID)
	}
	return nil
}
