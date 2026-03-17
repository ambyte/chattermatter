# ChatterMatter deployment

Deploy ChatterMatter from `ghcr.io/ambyte/chattermatter:latest` with PostgreSQL.

## Quick start

```bash
cd deploy
cp .env.example .env
# Edit .env - set DOMAIN, POSTGRES_PASSWORD, MM_SERVICESETTINGS_SITEURL

# Create volume directories
mkdir -p volumes/app/mattermost/{config,data,logs,plugins,client/plugins,bleve-indexes}
mkdir -p volumes/db/var/lib/postgresql/data

# Fix permissions (Mattermost runs as UID 2000)
sudo chown -R 2000:2000 volumes/app/mattermost

# For private image (ghcr.io): login first
docker login ghcr.io -u YOUR_GITHUB_USERNAME

docker compose up -d
```

## Configuration

- **DOMAIN** — hostname for SiteURL (e.g. `mattermost.example.com`)
- **POSTGRES_PASSWORD** — change from default
- **MM_SERVICESETTINGS_SITEURL** — full URL (e.g. `https://mattermost.example.com`)

Mattermost will be available at `http://localhost:8065` (or `APP_PORT` from .env).

## Troubleshooting

**"permission denied" on config.json** — fix volume ownership:
```bash
sudo chown -R 2000:2000 volumes/app/mattermost
docker compose restart mattermost
```
