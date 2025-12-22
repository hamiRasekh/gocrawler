# راهنمای راه‌اندازی Docker

## ساخت و اجرای Containers

### 1. ساخت و اجرا با Docker Compose

```bash
# ساخت و اجرای همه services
docker-compose up -d

# مشاهده لاگ‌ها
docker-compose logs -f crawler

# توقف services
docker-compose down

# ساخت مجدد images
docker-compose build --no-cache
docker-compose up -d
```

### 2. بررسی وضعیت Services

```bash
# لیست containers
docker-compose ps

# بررسی لاگ PostgreSQL
docker-compose logs postgres

# بررسی لاگ Crawler
docker-compose logs crawler
```

### 3. دسترسی به Container

```bash
# دسترسی به shell در container crawler
docker-compose exec crawler sh

# دسترسی به shell در container postgres
docker-compose exec postgres psql -U crawler -d crawler_db
```

### 4. اجرای Migrations

Migrations به صورت خودکار در startup اجرا می‌شوند. اما می‌توانید به صورت دستی هم اجرا کنید:

```bash
# دسترسی به database
docker-compose exec postgres psql -U crawler -d crawler_db

# سپس فایل‌های SQL را اجرا کنید
```

### 5. پاک کردن و شروع مجدد

```bash
# توقف و حذف containers, networks
docker-compose down

# حذف volumes (دیتابیس پاک می‌شود)
docker-compose down -v

# شروع مجدد
docker-compose up -d
```

## عیب‌یابی

### مشکل: Container start نمی‌شود

```bash
# بررسی لاگ‌ها
docker-compose logs crawler

# بررسی وضعیت container
docker-compose ps
```

### مشکل: اتصال به Database

```bash
# بررسی سلامت PostgreSQL
docker-compose exec postgres pg_isready -U crawler

# بررسی connection string
docker-compose exec crawler env | grep DB_
```

### مشکل: Chromium پیدا نمی‌شود

```bash
# بررسی نصب Chromium
docker-compose exec crawler which chromium-browser

# بررسی PATH
docker-compose exec crawler echo $CHROMIUM_PATH
```

### مشکل: Port در حال استفاده است

```bash
# تغییر port در docker-compose.yml
ports:
  - "8010:8009"  # به جای 8009:8009
```

## متغیرهای محیطی

متغیرهای محیطی در `docker-compose.yml` تنظیم شده‌اند. برای تغییر:

```yaml
environment:
  SERVER_PORT: 8009
  DB_HOST: postgres
  # ...
```

یا با `.env` file:

```bash
# ایجاد .env file
cp .env.example .env

# ویرایش .env
# سپس در docker-compose.yml از env_file استفاده کنید
```

## به‌روزرسانی Application

```bash
# Pull آخرین تغییرات
git pull

# ساخت مجدد image
docker-compose build crawler

# Restart container
docker-compose up -d crawler
```

## Cleanup

```bash
# حذف images استفاده نشده
docker image prune -a

# حذف همه containers
docker container prune

# حذف همه volumes
docker volume prune
```

