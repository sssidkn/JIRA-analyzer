# JIRA-analyzer

## Архитектура проекта v1.0

```
JIRA-analyzer/
├── JIRA-connector/
│   ├── proto/
│   ├── cmd/
│   │   └── service/
│   │       └── main.go         # Точка входа
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go       # Чтение конфигурации
│   │   ├── models/
│   │   │   ├── dto/
│   │   │   └── models.go       
│   │   ├── service/            # Бизнес логика
│   │   │   ├── service.go
│   │   │   └── connector.go
│   │   ├── repository/
│   │   │   └── repository.go   # TODO
│   │   └── transport/
│   │       ├── http/
│   │       │   ├── handlers/
│   │       │   ├── server.go   
│   │       └───└── router.go
│   ├── pkg/
│   │   ├── logger/
│   │   │   └── logger.go
│   │   └── db/
│   │       └── postgres/
│   │           └── postgres.go
│   └── go.mod
├── backend/
├── database/  
├── frontend/  
├── docker-compose.yml                    
└── README.md

```

## Схема TODO... И вообще все TODO