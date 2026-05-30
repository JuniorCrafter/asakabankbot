package repository

import (
	"database/sql"
	"encoding/json"
	"log"

	"asakabankbot/internal/domain"
)

type DepartmentRepository struct {
	db *sql.DB
}

func NewDepartmentRepository(db *sql.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// GetAllDepartments возвращает список всех отделов и их услуг из базы данных
func (r *DepartmentRepository) GetAllDepartments() ([]domain.Department, error) {
	query := `SELECT id, name, services FROM departments ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []domain.Department
	for rows.Next() {
		var d domain.Department
		var servicesJSON []byte

		if err := rows.Scan(&d.ID, &d.Name, &servicesJSON); err != nil {
			log.Printf("Ошибка при чтении отдела: %v", err)
			continue
		}

		if err := json.Unmarshal(servicesJSON, &d.Services); err != nil {
			log.Printf("Ошибка при парсинге услуг отдела %s: %v", d.Name, err)
		}

		deps = append(deps, d)
	}

	return deps, nil
}

// GetDepartmentByID ищет отдел и его услуги по ID
func (r *DepartmentRepository) GetDepartmentByID(id int) (*domain.Department, error) {
	query := `SELECT id, name, services FROM departments WHERE id = $1`

	var d domain.Department
	var servicesJSON []byte

	err := r.db.QueryRow(query, id).Scan(&d.ID, &d.Name, &servicesJSON)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(servicesJSON, &d.Services); err != nil {
		return nil, err
	}

	return &d, nil
}
