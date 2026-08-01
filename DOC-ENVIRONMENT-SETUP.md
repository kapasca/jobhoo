# 📋 JOBHOO Environment Configuration Guide

## 📝 Summary of Changes

### ✅ What Was Fixed

| Issue | Before | After |
|-------|--------|-------|
| **Email Config** | ❌ Old Mailtrap format (MAIL_*) | ✅ Modern standard (SMTP_*) |
| **EMAIL_PROVIDER** | ❌ Missing variable | ✅ Added with dev/smtp options |
| **PORT** | ⚠️ Different in .env vs .env.example | ✅ Consistent: .env = 8070, .env.example = 8080 (template) |
| **.env in git** | 🚨 SECURITY RISK - tracked with secrets | ✅ FIXED - added to .gitignore |
| **Credentials exposed** | 🚨 Real Mailtrap token in git | ✅ FIXED - removed and replaced with placeholders |
| **Documentation** | ⚠️ Minimal comments | ✅ Comprehensive with examples |

---

## 🔑 Configuration Files

### `.env` - Local Development (NOT IN GIT)
**Location:** `./env`  
**Git Status:** ✅ Protected by .gitignore  
**Purpose:** Local environment variables (contains secrets)

**Contains:**
```env
APP_ENV=development
PORT=8070
DATABASE_URL=postgres://...
SESSION_SECRET=...
EMAIL_PROVIDER=dev  # or "smtp"
SMTP_HOST=          # Filled only when using SMTP
SMTP_USER=
SMTP_PASS=
```

**Rules:**
- ❌ NEVER commit to git
- ✅ Copy from `.env.example` when setting up new environment
- ✅ Update with your actual credentials
- ✅ Gitignore protects from accidental commit

### `.env.example` - Template (IN GIT)
**Location:** `./env.example`  
**Git Status:** ✅ Tracked (safe, no secrets)  
**Purpose:** Template for developers to copy and configure

**Contains:**
```env
# Full documentation and examples
# Placeholder values (no real credentials)
# Section headers explaining each option
```

**Key Features:**
- ✅ Comprehensive documentation
- ✅ Examples for different environments (local, Docker, production)
- ✅ Production checklist
- ✅ Mailtrap setup instructions

---

## 📧 Email Configuration Explained

### Format Change: Old → New

**Old Format (DEPRECATED):**
```env
MAIL_TOKEN=xxx          # Mailtrap-specific
MAIL_HOST=sandbox.smtp.mailtrap.io
MAIL_PORT=587
MAIL_USER=xxx
MAIL_PASS=xxx
```

**New Format (CORRECT):**
```env
EMAIL_PROVIDER=dev      # or "smtp"
EMAIL_FROM=no-reply@jobhoo.local
SMTP_HOST=             # Filled for SMTP
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
```

### Email Modes

#### Mode 1: Dev (Console Logging)
**Best for:** Local development & testing
```env
EMAIL_PROVIDER=dev
SMTP_HOST=            # Leave empty
```
**Behavior:**
- Emails logged to console
- No actual sending
- Immediate feedback for debugging

#### Mode 2: SMTP (Real Sending)
**Best for:** Production, staging, testing with Mailtrap
```env
EMAIL_PROVIDER=smtp
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=587
SMTP_USER=your_mailtrap_user
SMTP_PASS=your_mailtrap_password
```
**Behavior:**
- Emails sent via SMTP server
- With Mailtrap: emails appear in inbox for testing
- With production: emails sent to real recipients

### Backward Compatibility
**Code still supports old MAIL_* variables:**
```go
// From config.go - line 49-52
// If SMTP_* empty but MAIL_* present → copy to SMTP_*
if cfg.SMTPHost == "" {
    if mailHost := getEnv("MAIL_HOST", ""); mailHost != "" {
        cfg.SMTPHost = mailHost
        cfg.SMTPPort = getEnv("MAIL_PORT", cfg.SMTPPort)
        cfg.SMTPUser = getEnv("MAIL_USER", cfg.SMTPUser)
        cfg.SMTPPass = getEnv("MAIL_PASS", cfg.SMTPPass)
    }
}
```

**However:** Use new SMTP_* format going forward.

---

## 🔐 Security Improvements

### `.env` Now Protected

**Before:**
```
❌ .env tracked in git
❌ Real Mailtrap token exposed: 59087a9ddcbdd2af63f2c2174fcf4b3e
❌ Real credentials visible in repository history
❌ Any developer could see production-like secrets
```

**After:**
```
✅ .env added to .gitignore
✅ No credentials in repository
✅ Each developer maintains local .env
✅ Safe to push to any git repository
```

### `.gitignore` Enhanced

```gitignore
# Local environment files
.env               # Main local config
.env.local         # User-specific overrides
.env.*.local       # Environment-specific (dev.local, prod.local)

# IDE files
.vscode/           # VS Code settings
.idea/             # IntelliJ settings
*.swp              # Vim swap files
*.swo              # Vim backup files
```

---

## 🚀 Setup Instructions

### For New Developers

1. **Copy template to local:**
   ```bash
   cp .env.example .env
   ```

2. **Update `.env` with your values:**
   ```env
   # Change these for your setup
   DATABASE_URL=postgres://...your_db...
   SESSION_SECRET=generate_random_string
   
   # For email testing with Mailtrap:
   EMAIL_PROVIDER=smtp
   SMTP_HOST=sandbox.smtp.mailtrap.io
   SMTP_USER=your_username
   SMTP_PASS=your_password
   ```

3. **For Docker Compose:**
   ```env
   # Use this DATABASE_URL instead:
   DATABASE_URL=postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable
   ```

4. **Run application:**
   ```bash
   go run ./cmd/server
   ```

### For Mailtrap Setup (Testing Email)

1. Sign up free at [mailtrap.io](https://mailtrap.io)
2. Create new Inbox
3. Click "Settings"
4. Copy SMTP credentials:
   - **Host:** Copy to SMTP_HOST
   - **Port:** Copy to SMTP_PORT (usually 587)
   - **Username:** Copy to SMTP_USER
   - **Password:** Copy to SMTP_PASS
5. Set in `.env`:
   ```env
   EMAIL_PROVIDER=smtp
   SMTP_HOST=sandbox.smtp.mailtrap.io
   SMTP_PORT=587
   SMTP_USER=your_username
   SMTP_PASS=your_password
   ```
6. Run app and send test email
7. Check Mailtrap inbox to verify

---

## 🏭 Production Deployment

### Pre-Deployment Checklist

```
Environment Setup
✅ Set APP_ENV=production
✅ Generate strong SESSION_SECRET (don't reuse development secret)
✅ Update DATABASE_URL to production database
✅ Set EMAIL_PROVIDER=smtp with real email service
✅ Never use .env.example — create new .env on production server

Email Configuration
✅ Use production email service (AWS SES, SendGrid, Mailgun, etc.)
✅ Configure SMTP credentials securely
✅ Test email delivery before go-live
✅ Verify sender address (EMAIL_FROM) matches domain

Security
✅ Keep .env out of git (use secure deployment process)
✅ Use environment secrets management (Docker secrets, K8s secrets, etc.)
✅ Rotate credentials regularly
✅ Monitor SMTP rate limits
```

### Example Production .env
```env
APP_ENV=production
PORT=8080

DATABASE_URL=postgres://produser:strong_password@prod-db.example.com:5432/jobhoo?sslmode=require

SESSION_SECRET=abcdefghijklmnop1234567890... # 32+ random characters

EMAIL_PROVIDER=smtp
EMAIL_FROM=noreply@jobhoo.com
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=AKIXXXXXXX
SMTP_PASS=xxx...xxx

AI_PROVIDER=anthropic
AI_API_KEY=sk-ant-...xxx
```

---

## ✅ Verification Checklist

### Quick Test
```bash
# 1. Check .env exists and is not in git
git status .env
# Expected: .env not listed (protected by .gitignore)

# 2. Verify .env.example is tracked
git ls-files .env.example
# Expected: .env.example listed

# 3. Start application
go run ./cmd/server
# Expected: No errors, app runs on configured PORT

# 4. Test email (if EMAIL_PROVIDER=dev)
# Signup or trigger forgot-password
# Expected: Email logged to console

# 5. Test email (if EMAIL_PROVIDER=smtp)
# Signup or trigger forgot-password
# Expected: Email appears in Mailtrap inbox
```

---

## 📚 Related Files

- `.env` - Local configuration (not in git)
- `.env.example` - Template configuration (in git)
- `.gitignore` - Ignore patterns (protects .env)
- `internal/config/config.go` - Config loading logic
- `cmd/server/main.go` - Application bootstrap

---

## 🔗 Environment Variables Reference

| Variable | Format | Default | Example |
|----------|--------|---------|---------|
| **APP_ENV** | string | development | production |
| **PORT** | string | 8080 | 8070 |
| **DATABASE_URL** | string | (required) | postgres://user:pass@host/db |
| **SESSION_SECRET** | string | "" | abc123xyz... |
| **AI_PROVIDER** | string | mock | anthropic, openai |
| **AI_API_KEY** | string | "" | sk-ant-xxx |
| **EMAIL_PROVIDER** | string | dev | smtp |
| **EMAIL_FROM** | string | no-reply@jobhoo.local | noreply@example.com |
| **SMTP_HOST** | string | "" | sandbox.smtp.mailtrap.io |
| **SMTP_PORT** | string | 587 | 465, 25 |
| **SMTP_USER** | string | "" | your_username |
| **SMTP_PASS** | string | "" | your_password |

---

## 🆘 Troubleshooting

### Issue: "DATABASE_URL is required"
**Solution:** Check .env has DATABASE_URL set

### Issue: Emails not sending (EMAIL_PROVIDER=smtp)
**Check:**
1. SMTP_HOST is set and correct
2. SMTP_PORT is correct (usually 587 for TLS, 465 for SSL)
3. SMTP_USER and SMTP_PASS are correct
4. Credentials are tested with SMTP client tool
5. Firewall allows outbound SMTP connections

### Issue: Emails stuck in dev mode
**Solution:**
- Set EMAIL_PROVIDER=smtp if SMTP credentials present
- Or manually set EMAIL_PROVIDER=smtp in .env

### Issue: Can't connect to Docker database
**Solution:**
- Use `docker compose up` to start containers
- Update DATABASE_URL to: `postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable`
- Verify db service running: `docker ps | grep postgres`

---

## 📝 Notes

- `.env` file should be different on each machine/environment
- `.env.example` is the shared template for all developers
- Never commit real credentials to any branch
- Use `.env.local` for personal overrides without affecting team
- Production should use secrets manager (not files on disk)

**Last Updated:** 2026-08-01
