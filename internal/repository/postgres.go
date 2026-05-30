package repository

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB(dbURL string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Ошибка драйвера БД: %v", err)
	}

	// Пытаемся достучаться до БД 5 раз с интервалом в 3 секунды
	for i := 1; i <= 5; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Успешное подключение к PostgreSQL!")
			return db
		}
		log.Printf("Попытка %d: База данных еще не готова (%v). Ждем 3 секунды...", i, err)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("Критическая ошибка: PostgreSQL недоступен после 5 попыток.")
	return nil
}
