# JIRA-analyzer

## Архитектура проекта v1.0

```
backend/
├── JIRA-connector/
│   ├── cmd/
│   │   └── service/
│   │       └── main.go         # Точка входа
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go       # Чтение конфигурации
│   │   ├── models/
│   │   │   ├── dto/
│   │   │   └── models.go       
│   │   ├── service/            # Бизнес-логика
│   │   │   └── connector.go
│   │   ├── repository/
│   │   │   └── repository.go   # TODO
│   │   └── transport/
│   │       ├── http/
│   │       │   ├── handlers/
│   │       │   ├── middleware/
│   │       │   │   └── logging.go
│   │       │   ├── server.go   # maybe ?
│   │       └───└── router.go
│   ├── pkg/
│   │   ├── logger/
│   │   │   └── logger.go
│   │   └── db/
│   │       └── postgres/
│   │           └── postgres.go
│   └── go.mod
├── frontend/                   # maybe
└── README.md

```

## Схема TODO... И вообще все TODO