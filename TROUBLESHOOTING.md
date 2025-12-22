# راهنمای عیب‌یابی Docker Build

## مشکلات رایج و راه‌حل‌ها

### مشکل 1: Go Dependencies Download Timeout

**خطا:**
```
go: golang.org/x/net@v0.19.0: unrecognized import path "golang.org/x/net": https fetch: Get "https://golang.org/x/net?go-get=1": net/http: TLS handshake timeout
```

**راه‌حل‌ها:**

1. **استفاده از VPN یا Proxy:**
   ```powershell
   # تنظیم proxy برای Docker (در صورت نیاز)
   $env:HTTP_PROXY="http://your-proxy:port"
   $env:HTTPS_PROXY="http://your-proxy:port"
   ```

2. **استفاده از Go Proxy ایرانی/چینی:**
   - Dockerfile به صورت خودکار از `goproxy.cn` و `goproxy.io` استفاده می‌کند
   - اگر باز هم مشکل داشت، می‌توانید به صورت دستی تنظیم کنید

3. **Build در زمان دیگری:**
   - گاهی اوقات مشکل از سمت سرور Go است
   - چند ساعت بعد دوباره امتحان کنید

### مشکل 2: Alpine Repository Connection Failed

**خطا:**
```
WARNING: fetching https://dl-cdn.alpinelinux.org/alpine/v3.22/main: could not connect to server
ERROR: unable to select packages
```

**راه‌حل‌ها:**

1. **استفاده از Mirror:**
   - Dockerfile به صورت خودکار از mirror استفاده می‌کند
   - اگر باز هم مشکل داشت، می‌توانید به صورت دستی تنظیم کنید

2. **بررسی اتصال اینترنت:**
   ```powershell
   # تست اتصال به Alpine CDN
   Test-NetConnection dl-cdn.alpinelinux.org -Port 443
   
   # تست اتصال به Mirror
   Test-NetConnection mirror.alpinelinux.org -Port 443
   ```

3. **استفاده از VPN:**
   - اگر در ایران هستید، ممکن است نیاز به VPN داشته باشید

### مشکل 3: Docker Build خیلی کند است

**راه‌حل‌ها:**

1. **استفاده از Build Cache:**
   ```powershell
   # Build با cache
   docker-compose build
   
   # Build بدون cache (فقط در صورت نیاز)
   docker-compose build --no-cache
   ```

2. **استفاده از Arvan Docker Registry:**
   - تنظیمات Docker Desktop را طبق `SETUP_ARVAN_WINDOWS.md` انجام دهید
   - این کار سرعت دانلود images را بسیار افزایش می‌دهد

### مشکل 4: Chromium نصب نمی‌شود

**خطا:**
```
chromium (no such package)
```

**راه‌حل:**

1. **بررسی Alpine Version:**
   - مطمئن شوید که از Alpine 3.22 یا بالاتر استفاده می‌کنید
   - Dockerfile به صورت خودکار version را تشخیص می‌دهد

2. **استفاده از Repository Community:**
   - Chromium در community repository است
   - Dockerfile به صورت خودکار آن را اضافه می‌کند

### مشکل 5: Build موفق می‌شود اما Container Start نمی‌شود

**راه‌حل‌ها:**

1. **بررسی لاگ‌ها:**
   ```powershell
   docker-compose logs crawler
   docker-compose logs postgres
   ```

2. **بررسی Health Check:**
   ```powershell
   docker-compose ps
   ```

3. **اجرای دستی Container:**
   ```powershell
   docker-compose run --rm crawler sh
   ```

### مشکل 6: PostgreSQL Connection Failed

**خطا:**
```
failed to connect to database
```

**راه‌حل‌ها:**

1. **بررسی که PostgreSQL Container Running است:**
   ```powershell
   docker-compose ps postgres
   ```

2. **بررسی Environment Variables:**
   ```powershell
   docker-compose exec crawler env | grep DB_
   ```

3. **تست Connection:**
   ```powershell
   docker-compose exec postgres psql -U crawler -d crawler_db -c "SELECT 1;"
   ```

### مشکل 7: Migrations اجرا نمی‌شوند

**راه‌حل‌ها:**

1. **اجرای دستی Migrations:**
   ```powershell
   # دسترسی به PostgreSQL
   docker-compose exec postgres psql -U crawler -d crawler_db
   
   # سپس فایل‌های SQL را اجرا کنید
   \i /docker-entrypoint-initdb.d/001_initial.up.sql
   ```

2. **بررسی فایل‌های Migration:**
   ```powershell
   docker-compose exec crawler ls -la /root/migrations
   ```

## راه‌حل‌های عمومی

### پاک کردن Cache و Build مجدد

```powershell
# پاک کردن همه containers
docker-compose down

# پاک کردن volumes (دقت: داده‌ها پاک می‌شوند)
docker-compose down -v

# پاک کردن build cache
docker builder prune -a

# Build مجدد
docker-compose build --no-cache
docker-compose up -d
```

### بررسی Network

```powershell
# بررسی network Docker
docker network ls

# بررسی network crawler-network
docker network inspect embroidery-designs_crawler-network
```

### بررسی Resource Usage

```powershell
# بررسی استفاده از منابع
docker stats

# بررسی disk usage
docker system df
```

## درخواست کمک

اگر هیچ‌کدام از راه‌حل‌ها کار نکرد:

1. **جمع‌آوری اطلاعات:**
   ```powershell
   # لاگ‌های کامل
   docker-compose logs > logs.txt
   
   # وضعیت containers
   docker-compose ps > status.txt
   
   # Docker info
   docker info > docker-info.txt
   ```

2. **ارسال اطلاعات:**
   - لاگ‌ها را بررسی کنید
   - خطای دقیق را پیدا کنید
   - متن خطا را برای بررسی ارسال کنید

## نکات مهم

- ✅ همیشه قبل از build، اتصال اینترنت را بررسی کنید
- ✅ در صورت مشکل، از VPN استفاده کنید
- ✅ تنظیمات Arvan Docker Registry را اعمال کنید (سرعت بیشتر)
- ✅ Build را در زمان‌های مختلف امتحان کنید (گاهی مشکل از سمت سرور است)
- ✅ از `docker-compose logs` برای بررسی مشکلات استفاده کنید

