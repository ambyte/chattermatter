# План удаления системы лицензирования

## Цель

Заменить систему лицензирования заглушками, которые всегда возвращают положительный результат, чтобы все функции были доступны без лицензии.

## Обнаруженные зависимости от лицензии

### Основные проверки лицензии

1. Проверки в конфигурации клиента (`config/client.go`)
2. Проверки в аутентификации (`app/authentication.go`)
3. Проверки в комплаенсе (`app/compliance.go`)
4. Проверки в LDAP (`app/ldap.go`)
5. Проверки в уведомлениях (`app/notification.go`)
6. Проверки в плагинах (`app/plugin_api.go`)
7. Проверки в поиске (Elasticsearch)
8. Проверки в кластере
9. Проверки в логировании
10. Проверки в ограничениях пользователей

### Структура Features

Все поля типа `*bool` в структуре `Features` используются для проверки наличия соответствующей лицензии:

- LDAP
- LDAPGroups
- MFA
- GoogleOAuth
- Office365OAuth
- OpenId
- Compliance
- Cluster
- Metrics
- MHPNS
- SAML
- Elasticsearch
- Announcement
- ThemeManagement
- EmailNotificationContents
- DataRetention
- MessageExport
- CustomPermissionsSchemes
- CustomTermsOfService
- GuestAccounts
- GuestAccountsPermissions
- IDLoadedPushNotifications
- LockTeammateNameDisplay
- EnterprisePlugins
- AdvancedLogging
- Cloud
- SharedChannels
- RemoteClusterService
- OutgoingOAuthConnections
- AutoTranslation
- FutureFeatures

## Подход к замене

### Вариант 1: Замена функций получения лицензии

Заменить методы получения лицензии в приложении на возврат заглушки с включенными всеми возможностями.

### Вариант 2: Модификация существующих проверок

Изменить каждую проверку лицензии, чтобы она всегда возвращала положительный результат.

### Рекомендуемый подход

Реализовать вариант 1, так как он централизованно решает проблему и требует минимальных изменений в остальном коде.

## Реализация заглушки

```go
func MockLicense() *model.License {
    license := &model.License{}
    license.Features = &model.Features{}

    // Устанавливаем все фичи в true для обеспечения работы всех возможностей
    license.Features.SetDefaults()

    // Перезаписываем значения, чтобы все возможности были активны
    trueVal := true
    userCount := 1000000 // Большое число пользователей

    license.Features.Users = &userCount
    license.Features.LDAP = &trueVal
    license.Features.LDAPGroups = &trueVal
    // ... и так далее для всех фич
    license.Features.FutureFeatures = &trueVal

    // Устанавливаем SKU на Enterprise для максимальной совместимости
    license.SkuShortName = model.LicenseShortSkuEnterpriseAdvanced

    return license
}
```

## Места, где нужно заменить получение лицензии

1. `server/platform/service.go` - метод `License()`
2. `server/channels/app/channels.go` - метод `License()`
3. `server/channels/app/server.go` - метод `License()`
4. Все места, где вызывается `app.License()` или `srv.License()`

## Тестирование

После замены необходимо протестировать:

1. Работу всех функций, которые ранее зависели от лицензии
2. Запуск сервера с новыми заглушками
3. Работу API-методов, которые проверяли лицензию
4. Работу админ-панели и настроек
