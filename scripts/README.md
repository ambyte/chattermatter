# Scripts

## generate_license.py

Генератор RSA-ключей и подписанных лицензий Mattermost. Совместим с валидацией лицензий Mattermost (RSA 2048, SHA512, PKCS1v15).

### Требования

```bash
pip install -r requirements.txt
```

### Использование

```bash
# Полная генерация: ключи + лицензия
python generate_license.py

# Только генерация ключей
python generate_license.py --keys-only

# Только лицензия (используя существующий приватный ключ)
python generate_license.py --license-only --private-key license-private.pem

# С параметрами
python generate_license.py --output-dir ./output --users 5000 --company "Acme Inc" --email admin@acme.com --days 365 --sku Enterprise

python generate_license.py --output-dir ./output --users 5000 --company "Apex" --email admin@apex-project.ru --days 900 --sku Enterprise
python generate_license.py  --license-only --private-key license-private.pem --output-dir ./output --users 5000 --company "Apex" --email admin@apex-project.ru --days 900 --sku Enterprise
```

### Параметры

| Параметр | По умолчанию | Описание |
|----------|--------------|----------|
| `--output-dir` | `.` | Папка для выходных файлов |
| `--keys-only` | — | Только сгенерировать ключи |
| `--license-only` | — | Только создать лицензию |
| `--private-key` | `license-private.pem` | Путь к приватному ключу |
| `--users` | `20000` | Количество лицензированных пользователей |
| `--company` | `My Company` | Название компании |
| `--email` | `admin@example.com` | Email администратора |
| `--days` | `365` | Срок действия лицензии (дней) |
| `--sku` | `Professional` | SKU: Professional, Enterprise, E10, E20 |

### Выходные файлы

| Файл | Описание |
|------|----------|
| `license-private.pem` | Приватный ключ (хранить в секрете) |
| `license-public-key.txt` | Публичный ключ в формате PEM |
| `mattermost.mattermost-license` | Подписанная лицензия (base64) |

### Использование с Mattermost (test environment)

1. Скопируйте `license-public-key.txt` в `server/channels/utils/license-public-key-test.txt`
2. Установите переменную окружения: `MM_SERVICEENVIRONMENT=test`
3. Поместите `mattermost.mattermost-license` в папку `config/` или загрузите через веб-интерфейс
