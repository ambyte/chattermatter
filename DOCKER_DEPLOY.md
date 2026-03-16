# Docker Deployment Guide

## GitHub Action - Build and Push Docker Image

### Описание

GitHub Action автоматически собирает и публикует Docker образ в GitHub Container Registry (GHCR) при пуше тега `deploy`.

### Файл workflow

`.github/workflows/deploy-docker.yml`

### Как запустить сборку

```bash
# Создать и запушить тег deploy
git tag -f deploy
git push origin deploy --force
```

### Что делает Action

1. Собирает Docker образ для платформ `linux/amd64` и `linux/arm64`
2. Публикует образ в GHCR с тегами:
    - `latest`
    - SHA коммита (короткий)
3. Использует кэширование для ускорения сборки

### Необходимые разрешения

Убедитесь, что в репозитории включены разрешения для GitHub Actions:

- Settings → Actions → General → Workflow permissions
- Выберите "Read and write permissions"

## Docker Compose Deployment

### Файл конфигурации

`docker-compose.ghcr.yml`

### Сервисы

#### 1. Mattermost Server

- **Образ**: `ghcr.io/{repository}/mattermost-server:latest`
- **Порты**:
    - `8065` - основной веб-интерфейс
    - `8067` - локальный API
- **Особенности**:
    - Лицензирование отключено (`DISABLE_LICENSE=1`)
    - Все enterprise-функции доступны

#### 2. PostgreSQL

- **Образ**: `postgres:15-alpine`
- **База данных**: `mattermost`
- **Пользователь**: `mmuser`
- **Пароль**: `mmuser-password`

#### 3. MinIO (S3-совместимое хранилище)

- **Образ**: `minio/minio:latest`
- **Порты**:
    - `9000` - API
    - `9001` - Web консоль
- **Bucket**: `mattermost` (создается автоматически)

### Быстрый старт

```bash
# 1. Установить переменную окружения с именем репозитория
export GITHUB_REPOSITORY=your-username/your-repo-name

# 2. Запустить все сервисы
docker-compose -f docker-compose.ghcr.yml up -d

# 3. Проверить статус
docker-compose -f docker-compose.ghcr.yml ps

# 4. Просмотр логов
docker-compose -f docker-compose.ghcr.yml logs -f mattermost
```

### Доступ к сервисам

| Сервис        | URL                   | Учетные данные                   |
| ------------- | --------------------- | -------------------------------- |
| Mattermost    | http://localhost:8065 | Создайте первого пользователя    |
| MinIO Console | http://localhost:9001 | minioadmin / minioadmin-password |

### Остановка

```bash
# Остановить сервисы
docker-compose -f docker-compose.ghcr.yml down

# Остановить и удалить volumes (все данные будут потеряны!)
docker-compose -f docker-compose.ghcr.yml down -v
```

### Переменные окружения

#### Mattermost

- `MM_SQLSETTINGS_DRIVERNAME` - драйвер БД (postgres)
- `MM_SQLSETTINGS_DATASOURCE` - строка подключения к БД
- `MM_FILESETTINGS_DRIVERNAME` - драйвер файлового хранилища (amazons3)
- `MM_FILESETTINGS_AMAZONS3ENDPOINT` - endpoint MinIO
- `DISABLE_LICENSE` - отключение проверки лицензии (1)

#### PostgreSQL

- `POSTGRES_USER` - пользователь БД
- `POSTGRES_PASSWORD` - пароль пользователя БД
- `POSTGRES_DB` - имя базы данных

#### MinIO

- `MINIO_ROOT_USER` - root пользователь
- `MINIO_ROOT_PASSWORD` - root пароль

### Volumes

Все данные сохраняются в Docker volumes:

- `mattermost-data` - данные Mattermost
- `mattermost-logs` - логи
- `mattermost-plugins` - плагины сервера
- `mattermost-client-plugins` - плагины клиента
- `postgres-data` - данные PostgreSQL
- `minio-data` - данные MinIO

### Безопасность

⚠️ **Важно**: Эта конфигурация предназначена для разработки и тестирования.

Для production использования:

1. Измените все пароли по умолчанию
2. Используйте SSL/TLS
3. Настройте firewall
4. Регулярно делайте бэкапы

### Устранение неполадок

#### Проверка здоровья сервисов

```bash
docker-compose -f docker-compose.ghcr.yml exec mattermost wget -qO- http://localhost:8065/api/v4/system/ping
```

#### Пересоздание с чистого листа

```bash
docker-compose -f docker-compose.ghcr.yml down -v
docker-compose -f docker-compose.ghcr.yml up -d
```

#### Просмотр логов конкретного сервиса

```bash
docker-compose -f docker-compose.ghcr.yml logs postgres
docker-compose -f docker-compose.ghcr.yml logs minio
docker-compose -f docker-compose.ghcr.yml logs mattermost
```
