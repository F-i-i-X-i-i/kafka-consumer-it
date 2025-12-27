# Kafka Consumer для обработки AI-команд на изображениях

Go-сервис, который подключается к Apache Kafka и получает сообщения с командами на обработку изображений с помощью ИИ.

## Возможности

- Подключение к Kafka как consumer group
- Десериализация JSON-сообщений
- Обработка команд: resize, filter, transform, analyze
- Graceful shutdown по SIGINT/SIGTERM
- Конфигурация через переменные окружения

## Формат сообщений

```json
{
  "id": "uuid-команды",
  "command": "resize|filter|transform|analyze",
  "image_url": "https://example.com/image.jpg",
  "parameters": {
    "width": 100,
    "height": 100
  }
}
```

## Быстрый старт

### 1. Запуск Kafka (через Docker)

```bash
docker-compose up -d
```

Дождитесь создания топика (около 15 секунд).

### 2. Запуск consumer

```bash
go run ./cmd/consumer
```

### 3. Отправка тестового сообщения

В отдельном терминале:

```bash
docker exec -it kafka kafka-console-producer --bootstrap-server localhost:9092 --topic image-commands
```

Введите JSON:
```json
{"id":"test-1","command":"resize","image_url":"http://example.com/img.jpg","parameters":{"width":100,"height":100}}
```

## Конфигурация

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `KAFKA_BROKERS` | Адреса Kafka брокеров (через запятую) | `localhost:9092` |
| `KAFKA_TOPIC` | Топик для чтения | `image-commands` |
| `KAFKA_GROUP_ID` | Consumer group ID | `image-processor-group` |

## Тестирование

```bash
go test ./... -v
```

## Структура проекта

```
├── cmd/consumer/main.go       # Точка входа
├── internal/
│   ├── config/config.go       # Конфигурация
│   ├── kafka/consumer.go      # Kafka consumer
│   ├── models/message.go      # Модели данных
│   └── processor/             # Обработчик команд + тесты
├── docker-compose.yml         # Kafka + Zookeeper
└── README.md
```
