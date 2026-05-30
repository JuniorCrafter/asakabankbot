# Этап 1: Сборка бинарного файла
FROM golang:alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем приложение (CGO_ENABLED=0 делает бинарник независимым от библиотек ОС)
RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot/main.go

# Этап 2: Создание минимального рабочего образа
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный файл из первого этапа
COPY --from=builder /app/bot .

# Копируем переменные окружения и папку с миграциями
COPY .env .
COPY migrations/ ./migrations/

# Указываем команду для запуска контейнера
CMD ["./bot"]