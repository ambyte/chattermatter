# Итоговый результат разговора: анализ и реализация подмены OpenID в Mattermost

## 1) Цель запроса

Была поставлена задача:

- проследить end-to-end flow авторизации через OpenID/OAuth;
- разобрать роль `OpenIdSettings`;
- оценить возможность подмены `github.com/mattermost/enterprise/oauth/openid` своим пакетом;
- затем перейти от анализа к реализации.

## 2) Что было проанализировано

### Ключевые точки flow (backend)

- Маршруты OAuth client flow в [`server/channels/web/oauth.go`](../server/channels/web/oauth.go).
- Основная OAuth/SSO-логика в [`server/channels/app/oauth.go`](../server/channels/app/oauth.go).
- User login error masking и проверка включенности SSO/OpenID в [`server/channels/api4/user.go`](../server/channels/api4/user.go).
- Сессионная логика логина в [`server/channels/app/login.go`](../server/channels/app/login.go).

### Контракт провайдера

- Реестр провайдеров и интерфейс в [`server/einterfaces/oauthproviders.go`](../server/einterfaces/oauthproviders.go).
- Генерируемый mock контракта в [`server/einterfaces/mocks/OAuthProvider.go`](../server/einterfaces/mocks/OAuthProvider.go).

### Конфиг OpenID

- Клиентские флаги/поля (`EnableSignUpWithOpenId`, `OpenIdButtonText`, `OpenIdButtonColor`) в [`server/config/client.go`](../server/config/client.go).
- Структура `SSOSettings`/`OpenIdSettings` в [`server/public/model/config.go`](../server/public/model/config.go).

### Механика enterprise-сборки

- Переменные и теги сборки в [`server/Makefile`](../server/Makefile).
- Build-tag импортов:
    - [`//go:build enterprise`](../server/enterprise/external_imports.go:4) в [`server/enterprise/external_imports.go`](../server/enterprise/external_imports.go);
    - [`//go:build enterprise || sourceavailable`](../server/enterprise/local_imports.go:4) в [`server/enterprise/local_imports.go`](../server/enterprise/local_imports.go).
- Импорт enterprise-пакета из main в [`server/cmd/mattermost/main.go`](../server/cmd/mattermost/main.go).

## 3) Архитектурный вывод

Подмена возможна, потому что прикладной код работает через интерфейс [`OAuthProvider`](../server/einterfaces/oauthproviders.go:13) и lookup в реестре.

Критично:

- зарегистрировать провайдер в `init()` под ключом `openid`;
- реализовать методы интерфейса;
- корректно выдавать пользователя из `userinfo`/`id_token`.

## 4) Что было реализовано в коде

### 4.1. Создан локальный OIDC-провайдер

Добавлен файл:

- [`server/enterprise/oauth/openid/openid.go`](../server/enterprise/oauth/openid/openid.go)

Реализовано:

- регистрация провайдера в реестре (`init` + `RegisterOAuthProvider`);
- [`GetSSOSettings`](../server/enterprise/oauth/openid/openid.go:35);
- [`GetUserFromIdToken`](../server/enterprise/oauth/openid/openid.go:44);
- [`GetUserFromJSON`](../server/enterprise/oauth/openid/openid.go:61);
- [`IsSameUser`](../server/enterprise/oauth/openid/openid.go:72);
- маппинг claims и fallback-логика пользователя.

### 4.2. Подключение импорта

Итоговое решение после обсуждения:

- **убрано** подключение локального openid из [`server/enterprise/external_imports.go`](../server/enterprise/external_imports.go);
- **добавлено** подключение локального openid в [`server/enterprise/local_imports.go`](../server/enterprise/local_imports.go).

Причина: локальный провайдер должен быть доступен и при `enterprise`, и при `sourceavailable`.

## 5) Важное уточнение по `BUILD_ENTERPRISE_READY`

Файлы `external_imports.go` / `local_imports.go` подключаются не "напрямую от переменной", а по build-tags.

- Переменная [`BUILD_ENTERPRISE_READY`](../server/Makefile:60) влияет на pipeline/ldflags и ветки Make;
- но включение конкретного файла определяет его `//go:build` условие и переданные `-tags`.

## 6) Проверки

Попытки тестового запуска сборки/пакета локального провайдера показали внешний блокер проекта:

- ошибки в [`server/einterfaces/autotranslation.go`](../server/einterfaces/autotranslation.go) по `model.Translation`.

Это не связано с OIDC-изменениями, а связано с текущим состоянием окружения/ветки.

## 7) Финальное состояние

- Локальный OIDC-провайдер добавлен: [`server/enterprise/oauth/openid/openid.go`](../server/enterprise/oauth/openid/openid.go).
- Подключен через source-available/enterprise импорты: [`server/enterprise/local_imports.go`](../server/enterprise/local_imports.go).
- В enterprise external imports локальная подмена больше не торчит: [`server/enterprise/external_imports.go`](../server/enterprise/external_imports.go).

## 8) Связанные документы

- Общий анализ enterprise импортов: [`plans/enterprise_external_imports_analysis.md`](enterprise_external_imports_analysis.md).
- План workflow enterprise сборки: [`plans/enterprise_build_workflow_plan.md`](enterprise_build_workflow_plan.md).
