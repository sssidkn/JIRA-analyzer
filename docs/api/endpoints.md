## `/api/v1/projects` (GET)

Получение всех загруженных проектов.

Тело ответа:

```json 
{
  "Projects": [
    {
      "Id": 0,
      "Key": "",
      "Name": "",
      "Url": ""
    }
  ]
}
```

## `/api/v1/projects/{id}` (GET)

Получение сухой статистики проекта по его ID в БД.

Тело ответа:

```json
{
  "Id": 0,
  "Key": "",
  "Name": "",
  "allIssuesCount": 0,
  "openIssuesCount": 0,
  "closeIssuesCount": 0,
  "resolvedIssuesCount": 0,
  "reopenedIssuesCount": 0,
  "progressIssuesCount": 0,
  "averageTime": 0,
  "averageIssuesCount": 0
}
```

## `/api/v1/projects/{id}` (DELETE)

Удаление проекта из БД по его ID.

## `/api/v1/connector/projects` (GET)

Получение списка доступных проектов из репозитория Jira.  
Параметры для пагинации и фильтрации:

- `limit` - количество проектов на странице.
- `page` - номер страницы, для которой мы хотим получить список проектов.
- `search` - параметр для фильтрации проектов по имени и ключу.

```json
{
  "Projects": [
    {
      "Id": 0,
      "Key": "",
      "Name": "",
      "Url": "",
      "Existence": false
    }
  ],
  "PageInfo": {
    "currentPage": 0,
    "pageCount": 0,
    "projectsCount": 0
  }
}
```

- `PageInfo` необходимо для того, чтобы пагинация в UI корректно работала.
- `currentPage` - номер страницы, для которой мы получили список проектов.
- `pageCount` - общее количество страниц, которое должно получиться при заданном параметре “limit”.
- `projectsCount` - общее количество доступных для загрузки проектов.

## `/api/v1/connector/updateProject/{projectKey}` (POST)

Обновление (или скачивание) проекта по его ключу.

## `/api/v1/graph/get/{taskNumber}` (GET)

Получение данных по аналитической задаче с номером taskNumber для проекта.

## `/api/v1/graph/make/{taskNumber}` (POST)

Проведение аналитической задачи с индексом taskNumber для проекта.

## `/api/v1/graph/delete` (DELETE)

Удаление всех аналитических задач для проекта.

## `/api/v1/isAnalyzed` (GET)

Получение информации о том, проведена ли хотя бы одна аналитическая задача для проекта.

## `/api/v1/compare/{taskNumber}` (GET)

Получение данных по аналитической задаче с индексом taskNumber для нескольких проектов.
