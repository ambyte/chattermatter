# Отключение системы лицензирования Mattermost

## Обзор

Система лицензирования Mattermost была модифицирована для работы без валидной лицензии. Теперь все enterprise-функции доступны по умолчанию.

## Как это работает

При установке переменной окружения `DISABLE_LICENSE=1`, система будет автоматически использовать заглушку лицензии со всеми включенными возможностями:

- LDAP/LDAP Groups
- SAML
- MFA
- Elasticsearch
- Clustering
- Metrics
- Compliance
- Data Retention
- Message Export
- Custom Permissions Schemes
- Guest Accounts
- Shared Channels
- Remote Cluster Service
- И многие другие...

## Запуск сервера без лицензии

### Linux/Mac:

```bash
export DISABLE_LICENSE=1
./mattermost
```

### Windows:

```cmd
set DISABLE_LICENSE=1
mattermost.exe
```

### Docker:

```yaml
environment:
    - DISABLE_LICENSE=1
```

## Файлы, которые были изменены

1. `server/channels/utils/license_mock.go` - создан файл с заглушкой лицензии
2. `server/channels/app/platform/license.go` - модифицированы методы `License()` и `SetLicense()`

## Структура заглушки

Заглушка создает полноценную лицензию с:

- SKU: Enterprise Advanced
- Лимит пользователей: 1,000,000
- Все фичи включены (true)

## Примечания

- Это изменение предназначено только для разработки и тестирования
- Для production-использования рекомендуется приобретать лицензию у Mattermost, Inc.
- Удаление лицензирования может нарушать условия лицензионного соглашения

## Технические детали

### Функция NewMockLicense()

```go
func NewMockLicense() *model.License
```

Создает лицензию со следующими характеристиками:

- Все булевы фичи установлены в `true`
- Лимит пользователей: 1,000,000
- SKU: Enterprise Advanced

### Переменная окружения

`DISABLE_LICENSE` - когда установлена (любое непустое значение), активирует использование заглушки вместо реальной лицензии.
