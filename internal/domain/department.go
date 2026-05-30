package domain

// Department описывает структуру банковского отдела
type Department struct {
	ID       int
	Name     string
	Services []string // Сюда мы будем распаковывать JSON из базы данных
}
