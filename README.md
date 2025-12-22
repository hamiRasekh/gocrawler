# 🕷️ Professional Go Crawler - سیستم کرالر حرفه‌ای

یک سیستم کرالر قدرتمند و حرفه‌ای با زبان Go که برای کرال کردن داده‌های طراحی‌های گلدوزی از وب‌سایت embroiderydesigns.com طراحی شده است. این پروژه از معماری Clean Architecture استفاده می‌کند و قابلیت‌های پیشرفته‌ای مانند headless browser، rotating proxy، و worker pool را ارائه می‌دهد.

## 📋 فهرست مطالب

- [ویژگی‌ها](#ویژگی‌ها)
- [معماری پروژه](#معماری-پروژه)
- [نحوه کار پروژه](#نحوه-کار-پروژه)
- [نصب و راه‌اندازی](#نصب-و-راه‌اندازی)
- [استفاده از API](#استفاده-از-api)
- [مستندات API](#مستندات-api)
- [پیکربندی](#پیکربندی)
- [توسعه](#توسعه)

## ✨ ویژگی‌ها

### قابلیت‌های اصلی

- ✅ **Headless Browser**: استفاده از Chromedp برای کرالینگ صفحات وب با JavaScript
- ✅ **Rotating Proxy**: مدیریت و چرخش خودکار proxy ها با health checking
- ✅ **Browser Fingerprinting**: شبیه‌سازی مرورگر واقعی با fingerprinting پیشرفته
- ✅ **REST API**: API کامل برای مدیریت tasks، proxies، و نتایج
- ✅ **PostgreSQL**: ذخیره‌سازی داده‌ها در PostgreSQL با migrations خودکار
- ✅ **Docker Support**: اجرای کامل در Docker و Docker Compose
- ✅ **Rate Limiting**: کنترل نرخ درخواست‌ها برای جلوگیری از بلاک شدن
- ✅ **Retry Mechanism**: تلاش مجدد خودکار با exponential backoff
- ✅ **Worker Pool**: پردازش همزمان با worker pool برای کارایی بالا
- ✅ **Embroidery API Crawler**: کرالر اختصاصی برای API سایت embroiderydesigns.com

### قابلیت‌های پیشرفته

- 🔄 **Auto-retry**: تلاش مجدد خودکار در صورت خطا
- 🎭 **Stealth Mode**: پنهان‌سازی هویت کرالر با تکنیک‌های پیشرفته
- 📊 **Real-time Monitoring**: نظارت بر وضعیت tasks و proxies
- 🗄️ **Product Management**: ذخیره‌سازی و مدیریت محصولات گلدوزی
- 🔍 **Advanced Filtering**: فیلتر پیشرفته محصولات بر اساس برند، کاتالوگ، و غیره

## 🏗️ معماری پروژه

پروژه با Clean Architecture و Dependency Injection طراحی شده است:

```
embroidery-designs/
├── cmd/
│   └── crawler/              # Entry point اصلی برنامه
│       └── main.go           # نقطه شروع برنامه
│
├── internal/
│   ├── api/                  # لایه API و HTTP handlers
│   │   ├── handlers.go       # Handler های REST API
│   │   ├── middleware.go    # Middleware ها (CORS, Logger, Recovery)
│   │   ├── routes.go        # تعریف route ها
│   │   └── server.go        # سرور HTTP
│   │
│   ├── browser/              # مدیریت مرورگر headless
│   │   ├── fingerprint.go   # تولید fingerprint مرورگر
│   │   ├── launcher.go      # راه‌اندازی مرورگر
│   │   ├── manager.go       # مدیریت مرورگر
│   │   └── stealth.go       # تکنیک‌های stealth
│   │
│   ├── config/               # مدیریت تنظیمات
│   │   └── config.go        # بارگذاری و مدیریت config
│   │
│   ├── crawler/              # هسته کرالر
│   │   ├── api_crawler.go          # کرالر REST API
│   │   ├── browser_crawler.go      # کرالر مرورگر
│   │   ├── embroidery_api_crawler.go  # کرالر اختصاصی embroidery
│   │   ├── web_crawler.go           # کرالر وب عمومی
│   │   ├── worker_pool.go          # Worker pool برای پردازش موازی
│   │   └── interface.go            # Interface های کرالر
│   │
│   ├── fingerprint/          # Browser fingerprinting
│   │   ├── headers.go       # تولید header های واقعی
│   │   └── profile.go       # پروفایل مرورگر
│   │
│   ├── proxy/                # مدیریت proxy
│   │   ├── health_checker.go # بررسی سلامت proxy
│   │   ├── manager.go       # مدیریت proxy ها
│   │   └── pool.go          # Pool مدیریت proxy
│   │
│   ├── service/              # لایه business logic
│   │   ├── crawler_service.go  # سرویس کرالر
│   │   └── task_service.go     # سرویس task
│   │
│   ├── storage/              # لایه دیتابیس
│   │   ├── models.go        # مدل‌های داده
│   │   ├── postgres.go      # اتصال PostgreSQL
│   │   └── repository.go    # Repository pattern
│   │
│   └── utils/                # ابزارهای کمکی
│       ├── logger.go        # Logger
│       ├── rate_limiter.go  # Rate limiter
│       └── retry.go         # Retry mechanism
│
├── migrations/               # Database migrations
│   ├── 001_initial.up.sql   # Migration اولیه
│   ├── 001_initial.down.sql
│   ├── 002_products.up.sql  # Migration محصولات
│   └── 002_products.down.sql
│
├── docker/                   # فایل‌های Docker
│   ├── Dockerfile           # Dockerfile اصلی
│   ├── Dockerfile.arvan     # Dockerfile برای Arvan Cloud
│   ├── entrypoint.sh        # اسکریپت راه‌اندازی
│   └── daemon.json.example  # تنظیمات Docker daemon
│
├── docker-compose.yml        # Docker Compose configuration
├── go.mod                    # Go modules
├── Makefile                  # دستورات Make
└── README.md                 # این فایل
```

## 🔄 نحوه کار پروژه

### جریان کلی کار

1. **راه‌اندازی**: برنامه با بارگذاری تنظیمات و اتصال به دیتابیس شروع می‌شود
2. **ایجاد Task**: کاربر از طریق API یک task جدید ایجاد می‌کند
3. **شروع Crawling**: با فراخوانی API، task شروع به اجرا می‌شود
4. **پردازش**: Worker pool task را دریافت کرده و با کرالر مناسب پردازش می‌کند
5. **ذخیره‌سازی**: نتایج در دیتابیس ذخیره می‌شوند
6. **نظارت**: وضعیت task و نتایج از طریق API قابل مشاهده است

### انواع کرالر

#### 1. API Crawler (`api_crawler.go`)
- برای کرال کردن REST API ها استفاده می‌شود
- از HTTP client با پشتیبانی proxy استفاده می‌کند
- Rate limiting و retry mechanism دارد

#### 2. Browser Crawler (`browser_crawler.go`)
- برای کرال کردن صفحات وب با JavaScript استفاده می‌شود
- از Chromedp برای کنترل headless browser استفاده می‌کند
- Stealth techniques برای پنهان‌سازی هویت

#### 3. Embroidery API Crawler (`embroidery_api_crawler.go`)
- کرالر اختصاصی برای API سایت embroiderydesigns.com
- پشتیبانی از pagination خودکار
- ذخیره‌سازی خودکار محصولات در دیتابیس
- پردازش و تبدیل داده‌های Elasticsearch

### Worker Pool

Worker Pool برای پردازش موازی tasks استفاده می‌شود:
- تعداد worker ها قابل تنظیم است (پیش‌فرض: 10)
- هر worker یک task را به صورت مستقل پردازش می‌کند
- Context cancellation برای توقف graceful

### Proxy Management

- **Health Checking**: بررسی خودکار سلامت proxy ها
- **Rotation**: چرخش خودکار proxy ها
- **Failure Tracking**: ردیابی proxy های ناموفق
- **Auto-disable**: غیرفعال کردن خودکار proxy های مشکل‌دار

### Browser Fingerprinting

- تولید User-Agent واقعی
- تنظیم Header های مرورگر
- Stealth techniques برای جلوگیری از تشخیص

## 🚀 نصب و راه‌اندازی

### پیش‌نیازها

- Go 1.21 یا بالاتر
- Docker & Docker Compose
- PostgreSQL 15+ (یا استفاده از Docker Compose)

### روش 1: با Docker Compose (توصیه می‌شود)

1. **کلون کردن پروژه:**
```bash
git clone <repository-url>
cd embroidery-designs
```

2. **اجرای بک‌اند (Go) با Docker Compose:**
```bash
docker compose up -d
```

این دستور فقط سرویس‌های Go (اپلیکیشن و PostgreSQL) را بالا می‌آورد. پس از اجرا، می‌توانید لاگ‌ها را با دستور زیر ببینید:
```bash
docker compose logs -f crawler
```

3. **اجرای فرانت‌اند React در کانتینر مجزا:**
```bash
docker compose -f docker-compose.frontend.yml up -d
```

این فایل Compose فقط UI را بالا می‌آورد و آن را به همان شبکه‌ی `production_network` متصل می‌کند. در صورت نیاز به توقف یا مشاهده‌ی لاگ‌ها:
```bash
docker compose -f docker-compose.frontend.yml logs -f frontend
docker compose -f docker-compose.frontend.yml down
```

بعد از بالا آمدن فرانت‌اند، صفحه‌ی «Crawler Config» از مسیر `http://localhost:3009/crawler/config` قابل دسترسی است و امکان ویرایش JSON مربوط به فیلترهای کرالر را فراهم می‌کند.

4. **بررسی وضعیت:**
```bash
curl http://localhost:8009/api/v1/health
```

### روش 2: بدون Docker

1. **نصب وابستگی‌ها:**
```bash
go mod download
```

2. **راه‌اندازی PostgreSQL:**
```bash
# راه‌اندازی PostgreSQL (مثال)
psql -U postgres -c "CREATE DATABASE crawler_db;"
```

3. **اجرای migrations:**
```bash
# نیاز به golang-migrate
migrate -path migrations -database "postgres://crawler:password@localhost:5432/crawler_db?sslmode=disable" up
```

4. **تنظیم متغیرهای محیطی:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=crawler
export DB_PASSWORD=password
export DB_NAME=crawler_db
```

5. **اجرای برنامه:**
```bash
go run cmd/crawler/main.go
```

## 📡 استفاده از API

### مشاهده Swagger UI

بعد از اجرای سرویس می‌توانید مستقیماً به آدرس [http://localhost:8009/swagger](http://localhost:8009/swagger) بروید تا نسخه تعاملی مستندات (Swagger UI) همراه با تست‌کننده آنلاین را مشاهده کنید. فایل `docs/swagger/openapi.yaml` منبع این مستندات است و در صورت نیاز می‌توانید آن را برای پیاده‌سازی‌های سفارشی ویرایش کنید.

### مثال 1: ایجاد Task برای کرال Embroidery API

```bash
curl -X POST http://localhost:8009/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Embroidery Products Crawl",
    "url": "https://www.embroiderydesigns.com/es/prdsrch",
    "type": "api",
    "config": {
      "crawler_type": "embroidery_api"
    }
  }'
```

### مثال 2: شروع Crawling

```bash
# جایگزین {task_id} با ID واقعی task
curl -X POST http://localhost:8009/api/v1/tasks/{task_id}/start
```

### مثال 3: دریافت نتایج

```bash
# دریافت لیست محصولات
curl http://localhost:8009/api/v1/tasks/{task_id}/results

# دریافت محصولات با فیلتر
curl "http://localhost:8009/api/v1/products?brand=ABC&in_stock=true&limit=20"
```

### مثال 4: مدیریت Proxy

```bash
# اضافه کردن proxy
curl -X POST http://localhost:8009/api/v1/proxies \
  -H "Content-Type: application/json" \
  -d '{
    "host": "proxy.example.com",
    "port": 8080,
    "type": "http",
    "username": "user",
    "password": "pass"
  }'

# لیست proxy ها
curl http://localhost:8009/api/v1/proxies
```

### مثال 5: شروع سریع کرال محصولات Embroidery

در صورت نیاز به راه‌اندازی سریع کرال اختصاصی سایت، تنها کافیست endpoint زیر را صدا بزنید. یک فایل نمونه برای ابزارهای `REST Client` / `Insomnia` در مسیر `docs/examples/embroidery-crawl.http` قرار داده شده است.

```bash
curl -X POST http://localhost:8009/api/v1/products/crawl
```

## 📚 مستندات API

نسخه کامل OpenAPI/Swagger با مسیر `/swagger` در دسترس است و امکان دانلود فایل `openapi.yaml` یا اتصال آن به محیط‌های خارجی (مثلاً SwaggerHub یا Postman) فراهم شده است.

### Auth

#### ثبت‌نام ادمین جدید
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "new_admin",
  "password": "StrongPass123"
}
```

#### ورود (Login)
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

#### صدور توکن API یک‌ساله
```http
POST /api/v1/auth/admin-token
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "token_name": "dashboard-bot"
}
```

### Tasks

#### ایجاد Task جدید
```http
POST /api/v1/tasks
Content-Type: application/json

{
  "name": "Example Task",
  "url": "https://example.com/api/data",
  "type": "api",
  "config": {
    "headers": {
      "Authorization": "Bearer token"
    },
    "crawler_type": "embroidery_api"  // برای کرالر embroidery
  }
}
```

#### لیست Tasks
```http
GET /api/v1/tasks?limit=10&offset=0
```

#### دریافت Task
```http
GET /api/v1/tasks/:id
```

#### آپدیت Task
```http
PUT /api/v1/tasks/:id
Content-Type: application/json

{
  "name": "Updated Name",
  "url": "https://new-url.com"
}
```

#### حذف Task
```http
DELETE /api/v1/tasks/:id
```

#### شروع Crawling
```http
POST /api/v1/tasks/:id/start
```

#### توقف Crawling
```http
POST /api/v1/tasks/:id/stop
```

#### Pause/Resume
```http
POST /api/v1/tasks/:id/pause
POST /api/v1/tasks/:id/resume
```

#### دریافت وضعیت Task
```http
GET /api/v1/tasks/:id/status
```

#### دریافت نتایج
```http
GET /api/v1/tasks/:id/results?limit=10&offset=0
```

#### حذف نتایج
```http
DELETE /api/v1/tasks/:id/results
```

### Products

#### لیست محصولات
```http
GET /api/v1/products?limit=20&offset=0&brand=ABC&in_stock=true
```

#### دریافت محصول
```http
GET /api/v1/products/:id
```

#### آمار محصولات
```http
GET /api/v1/products/stats
```

### Embroidery Crawl Config

#### دریافت تنظیمات فعلی
```http
GET /api/v1/products/crawl-config
```

#### بروزرسانی فیلترها و ورودی‌های API
```http
PUT /api/v1/products/crawl-config
Content-Type: application/json

{
  "payload_overrides": {
    "query": {
      "bool": {
        "must": [
          { "term": { "definitionName": "StockDesign" } },
          { "term": { "catalog.raw": "Christmas" } }
        ]
      }
    }
  }
}
```

> مقدار `payload_overrides` عیناً روی payload پایه کرالر مرج می‌شود. فیلدهای صفحه‌بندی (`from` و `size`) همیشه توسط سیستم مدیریت می‌شوند و نیازی به ارسال آن‌ها نیست. برای تجربه‌ای بهتر می‌توانید از صفحه **Crawler Config** در رابط کاربری استفاده کنید که یک JSON editor آماده برای این کار دارد.

### Proxies

#### لیست Proxies
```http
GET /api/v1/proxies
```

#### اضافه کردن Proxy
```http
POST /api/v1/proxies
Content-Type: application/json

{
  "host": "proxy.example.com",
  "port": 8080,
  "type": "http",
  "username": "user",
  "password": "pass"
}
```

#### حذف Proxy
```http
DELETE /api/v1/proxies/:id
```

#### تست Proxy
```http
POST /api/v1/proxies/test
Content-Type: application/json

{
  "host": "proxy.example.com",
  "port": 8080,
  "type": "http"
}
```

### System

#### Health Check
```http
GET /api/v1/health
```

#### آمار سیستم
```http
GET /api/v1/stats
```

## ⚙️ پیکربندی

متغیرهای محیطی قابل تنظیم:

### Server
- `SERVER_HOST`: آدرس سرور (پیش‌فرض: `0.0.0.0`)
- `SERVER_PORT`: پورت سرور (پیش‌فرض: `8009`)
- `API_PREFIX`: پیشوند API (پیش‌فرض: `/api/v1`)

### Database
- `DB_HOST`: آدرس دیتابیس (پیش‌فرض: `localhost`)
- `DB_PORT`: پورت دیتابیس (پیش‌فرض: `5432`)
- `DB_USER`: نام کاربری (پیش‌فرض: `crawler`)
- `DB_PASSWORD`: رمز عبور
- `DB_NAME`: نام دیتابیس (پیش‌فرض: `crawler_db`)
- `DB_SSLMODE`: حالت SSL (پیش‌فرض: `disable`)

### Logging
- `LOG_LEVEL`: سطح لاگ (پیش‌فرض: `info`) - مقادیر: `debug`, `info`, `warn`, `error`
- `LOG_FORMAT`: فرمت لاگ (پیش‌فرض: `json`) - مقادیر: `json`, `text`

### Crawler
- `MAX_WORKERS`: تعداد worker ها (پیش‌فرض: `10`)
- `RATE_LIMIT_PER_SECOND`: نرخ درخواست در ثانیه (پیش‌فرض: `5`)
- `REQUEST_TIMEOUT`: تایم‌اوت درخواست (پیش‌فرض: `30s`)
- `RETRY_MAX_ATTEMPTS`: تعداد تلاش‌های مجدد (پیش‌فرض: `3`)
- `RETRY_BACKOFF_MULTIPLIER`: ضریب backoff (پیش‌فرض: `2`)

### Browser
- `HEADLESS`: اجرای headless (پیش‌فرض: `true`)
- `BROWSER_TIMEOUT`: تایم‌اوت مرورگر (پیش‌فرض: `60s`)
- `USER_DATA_DIR`: مسیر داده‌های مرورگر (پیش‌فرض: `/tmp/browser-data`)

### Proxy
- `PROXY_ENABLED`: فعال/غیرفعال بودن proxy (پیش‌فرض: `true`)
- `PROXY_HEALTH_CHECK_INTERVAL`: فاصله بررسی سلامت (پیش‌فرض: `5m`)
- `PROXY_MAX_FAILURES`: حداکثر تعداد خطا قبل از غیرفعال شدن (پیش‌فرض: `3`)

### Auth
- `JWT_SECRET`: کلید امضای JWT (پیش‌فرض: مقدار نمونه‌ای که باید تغییر کند)
- `JWT_EXPIRATION`: مدت اعتبار دسترسی (پیش‌فرض: `24h`)
- `REFRESH_TOKEN_EXPIRATION`: مدت اعتبار refresh token (پیش‌فرض: `168h`)
- `ADMIN_TOKEN_LIFETIME`: طول عمر پیش‌فرض توکن‌های API صادر شده برای کاربر ادمین (پیش‌فرض: `8760h` یعنی یک سال)

## 🔧 توسعه

### ساخت پروژه
```bash
go build -o crawler ./cmd/crawler
```

### اجرای تست‌ها
```bash
go test ./...
```

### فرمت کردن کد
```bash
go fmt ./...
```

### اجرای Linter
```bash
golangci-lint run
```

### ساخت Docker Image
```bash
docker build -f docker/Dockerfile -t crawler:latest .
```

## 📊 ساختار دیتابیس

### Tables

1. **tasks**: ذخیره‌سازی tasks
   - id, name, url, type, status, config, created_at, updated_at, started_at, completed_at

2. **crawl_results**: نتایج کرال
   - id, task_id, url, method, status_code, headers, body, response_time, proxy_used, created_at

3. **proxies**: لیست proxy ها
   - id, host, port, type, username, password, is_active, failure_count, last_checked, created_at, updated_at

4. **crawl_logs**: لاگ‌های کرال
   - id, task_id, level, message, metadata, created_at

5. **products**: محصولات گلدوزی
   - id, elastic_id, product_id, item_id, name, brand, catalog, artist, rating, prices, stock info, categories, keywords, variants, raw_data, created_at, updated_at

## 🐛 عیب‌یابی

برای مشکلات رایج، به فایل `TROUBLESHOOTING.md` مراجعه کنید.

### مشکلات رایج

1. **خطای اتصال به دیتابیس**: بررسی کنید PostgreSQL در حال اجرا است و تنظیمات درست است
2. **خطای proxy**: بررسی کنید proxy ها معتبر هستند و health check فعال است
3. **خطای browser**: بررسی کنید Chromium نصب است و مسیر درست تنظیم شده است

## 📝 License

MIT

## 🤝 مشارکت

برای مشارکت در پروژه:
1. Fork کنید
2. Branch جدید ایجاد کنید (`git checkout -b feature/AmazingFeature`)
3. Commit کنید (`git commit -m 'Add some AmazingFeature'`)
4. Push کنید (`git push origin feature/AmazingFeature`)
5. Pull Request باز کنید

## 📞 پشتیبانی

برای سوالات و مشکلات، یک Issue در repository باز کنید.

---

**نکته**: این پروژه برای اهداف آموزشی و تحقیقاتی طراحی شده است. لطفاً قوانین و مقررات وب‌سایت‌های هدف را رعایت کنید.
