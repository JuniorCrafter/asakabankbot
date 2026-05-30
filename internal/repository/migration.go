package repository

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// RunMigrations проверяет и применяет новые SQL-файлы миграций
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// 1. Создаем служебную таблицу для отслеживания примененных миграций
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`)
	if err != nil {
		log.Printf("ОШИБКА: Не удалось создать таблицу миграций: %v", err)
		return err
	}

	// 2. Читаем все файлы в папке миграций
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Printf("ОШИБКА: Не удалось прочитать папку %s: %v", migrationsDir, err)
		return err
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	// Сортируем по имени (001_init.sql, 002_add_field.sql и т.д.)
	sort.Strings(migrationFiles)

	// 3. Выполняем те миграции, которых еще нет в базе
	for _, file := range migrationFiles {
		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, file).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			log.Printf("Применение миграции: %s...", file)

			content, err := os.ReadFile(filepath.Join(migrationsDir, file))
			if err != nil {
				return err
			}

			// Выполняем SQL-скрипт
			_, err = db.Exec(string(content))
			if err != nil {
				log.Printf("КРИТИЧЕСКАЯ ОШИБКА в миграции %s: %v", file, err)
				return err
			}

			// Записываем в историю, что миграция успешно применена
			_, err = db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, file)
			if err != nil {
				return err
			}
		}
	}

	log.Println("Миграции успешно проверены (БД в актуальном состоянии).")
	return nil
}
