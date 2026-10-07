# اجرای OUTSIDE در VS Code

## پیش‌نیازها
- Go نسخهٔ 1.23 یا جدیدتر
- Node.js نسخهٔ 20 یا جدیدتر
- PostgreSQL روشن و در دسترس. اگر Docker Desktop دارید، می‌توانید فرمان docker compose up -d را در پوشهٔ پروژه اجرا کنید.

## راه‌اندازی در ویندوز و PowerShell
1. فایل ZIP را استخراج کنید و پوشهٔ outside-storefront را در VS Code باز کنید.
2. ترمینال VS Code را در ریشهٔ پروژه باز کنید و فایل تنظیمات را بسازید:
   Copy-Item .env.example .env
3. فایل .env را باز کنید. مقدار JWT_SECRET را با یک رشتهٔ تصادفی حداقل ۳۲ کاراکتری جایگزین کنید. شمارهٔ مدیر را هم در صورت نیاز در ADMIN_PHONE بگذارید.
4. اگر PostgreSQL را با Docker اجرا نمی‌کنید، سرویس PostgreSQL را روشن کنید و مقادیر اتصال DATABASE_URL در .env را با تنظیمات خودتان هماهنگ کنید. دیتابیس shop باید وجود داشته باشد.
5. فرانت‌اند را بسازید:
   Set-Location .\frontend
   npm.cmd install
   npm.cmd run build
   Set-Location ..
6. از ریشهٔ پروژه سرور را اجرا کنید:
   go run ./cmd/api
7. سایت را در مرورگر باز کنید: http://localhost:8080

کد ورود یک‌بارمصرف تا زمان اتصال سرویس پیامک واقعی، در ترمینال سرور چاپ می‌شود. دیتابیس PostgreSQL باید هنگام اجرای برنامه در دسترس باشد.
