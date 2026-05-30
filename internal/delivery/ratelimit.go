package delivery

import (
	"sync"
	"time"
)

// RateLimiter ограничивает количество запросов от одного пользователя
type RateLimiter struct {
	visitors map[int64]*visitor
	mu       sync.Mutex
	rate     int           // Сколько запросов разрешено за окно
	window   time.Duration // Временное окно (например, 1 секунда)
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter создает новый лимитер
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[int64]*visitor),
		rate:     rate,
		window:   window,
	}

	// Запускаем фоновую очистку старых записей, чтобы не забивать память
	go rl.cleanup()
	return rl
}

// Allow проверяет, можно ли пропустить запрос от данного Telegram ID
func (rl *RateLimiter) Allow(tgID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[tgID]
	if !exists {
		rl.visitors[tgID] = &visitor{lastSeen: time.Now(), count: 1}
		return true
	}

	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	v.count++
	return v.count <= rl.rate
}

// cleanup удаляет пользователей, которые давно ничего не писали (раз в 5 минут)
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for tgID, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, tgID)
			}
		}
		rl.mu.Unlock()
	}
}
