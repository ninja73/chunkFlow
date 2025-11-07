# chunkFlow

`chunkFlow` — распределённый сервис хранения файлов, вдохновлённый подходом Amazon S3:
REST-интерфейс для клиентов + gRPC-взаимодействие между storage-нодами.

---

## Возможности

* Загрузка файла через REST (`POST /upload`)
* Скачивание файла (`GET /download/{fileID}`)
* Разбиение файла на чанки
* Распределение чанков по storage-нодам
* Межсерверное взаимодействие через gRPC

---

## Структура проекта

```
chunkFlow/
├── cmd/
│   ├── distributor/         # REST-сервер + координирует хранение/чтение
│   └── storage/             # отдельный storage-нод (хранит чанки)
├── internal/                # бизнес-сущности и интерфейсы
├── pkg/
│   └── proto/               # protobuf/gRPC API
├── tests/                   # интеграционные тесты (upload → download)
├── Makefile
├── distributor.Dockerfile
├── storage.Dockerfile
└── docker-compose.yaml
```

---

## Быстрый старт

### Запуск через Docker Compose

```bash
docker compose up --build
```

Будут подняты:

* `distributor` — REST API (`http://localhost:8080`)
* несколько storage-нод (`storage1`, `storage2`, ...)

---

### Загрузка файла

```bash
curl -X POST http://localhost:8080/upload \
  -H "Content-Type: multipart/form-data" \
  -F "file=@path/to/file.bin"
```

Ответ содержит `fileID`.

---

### Скачивание файла

```bash
curl http://localhost:8080/download/<fileID> -o result.bin
```

---

### Запуск тестов

```bash
go test ./tests -v
```

---

## Конфигурация

| Переменная      | Описание                                                                |
| --------------- | ----------------------------------------------------------------------- |
| `STORAGE_NODES` | список адресов storage-нод через запятую: `storage1:9001,storage2:9002` |

Данные теряются при перезапуске.

---

## Планы улучшений

* добавить базу для хранения методанных файла
* динамическое добавление нод хранения
* репликация чанков по нескольким нодам
* распределение через hashing
* метрики Prometheus / Grafana
* OpenAPI документация

---
