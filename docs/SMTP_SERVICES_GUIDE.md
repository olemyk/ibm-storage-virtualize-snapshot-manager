# SMTP Services Guide for Notifications

## Overview

This guide provides configuration details for popular SMTP services you can use with the Snapshot Manager notification system.

---

## 🔥 Recommended Services

### 1. Gmail (Google Workspace)

**Best For:** Small to medium deployments, testing

**Configuration:**

```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "username": "your-email@gmail.com",
  "password": "app-specific-password",
  "from_address": "your-email@gmail.com",
  "to_addresses": "recipient@company.com",
  "use_tls": true
}
```

**Setup Steps:**

1. Enable 2-Factor Authentication on your Google account
2. Go to: https://myaccount.google.com/apppasswords
3. Generate an "App Password" for "Mail"
4. Use the 16-character password (not your regular password)

**Limits:**
- 500 emails per day (free Gmail)
- 2,000 emails per day (Google Workspace)
- Rate limit: ~100 emails per hour

**Pros:**
- ✅ Easy to set up
- ✅ Reliable delivery
- ✅ Free for testing

**Cons:**
- ❌ Daily sending limits
- ❌ Requires app-specific password

---

### 2. Microsoft 365 / Outlook.com

**Best For:** Enterprise environments using Microsoft 365

**Configuration:**

```json
{
  "smtp_host": "smtp.office365.com",
  "smtp_port": 587,
  "username": "alerts@company.com",
  "password": "your-password",
  "from_address": "alerts@company.com",
  "to_addresses": "team@company.com",
  "use_tls": true
}
```

**Setup Steps:**

1. Use your Microsoft 365 email and password
2. Ensure SMTP AUTH is enabled (usually default)
3. For MFA accounts, use app password

**Limits:**
- 10,000 emails per day (Microsoft 365)
- 300 emails per day (Outlook.com free)
- Rate limit: 30 messages per minute

**Pros:**
- ✅ High sending limits
- ✅ Enterprise-grade reliability
- ✅ Good for corporate environments

**Cons:**
- ❌ Requires Microsoft 365 subscription for high limits
- ❌ May require app password with MFA

---

### 3. SendGrid

**Best For:** High-volume production environments

**Configuration:**

```json
{
  "smtp_host": "smtp.sendgrid.net",
  "smtp_port": 587,
  "username": "apikey",
  "password": "SG.your-api-key-here",
  "from_address": "noreply@yourdomain.com",
  "to_addresses": "alerts@company.com",
  "use_tls": true
}
```

**Setup Steps:**

1. Sign up at: https://sendgrid.com
2. Go to Settings → API Keys
3. Create new API key with "Mail Send" permission
4. Use "apikey" as username (literal string)
5. Use API key as password

**Limits:**
- 100 emails per day (free tier)
- 40,000+ emails per day (paid plans from $15/month)
- No rate limits on paid plans

**Pros:**
- ✅ Very high sending limits
- ✅ Excellent deliverability
- ✅ Detailed analytics
- ✅ No daily limits on paid plans

**Cons:**
- ❌ Requires account setup
- ❌ Free tier limited to 100/day

---

### 4. Mailgun

**Best For:** Developers, API-first approach

**Configuration:**

```json
{
  "smtp_host": "smtp.mailgun.org",
  "smtp_port": 587,
  "username": "postmaster@your-domain.mailgun.org",
  "password": "your-smtp-password",
  "from_address": "alerts@your-domain.mailgun.org",
  "to_addresses": "team@company.com",
  "use_tls": true
}
```

**Setup Steps:**

1. Sign up at: https://www.mailgun.com
2. Add and verify your domain
3. Get SMTP credentials from domain settings
4. Use provided username and password

**Limits:**
- 5,000 emails per month (free tier)
- 50,000+ emails per month (paid plans from $35/month)

**Pros:**
- ✅ Developer-friendly
- ✅ Good free tier
- ✅ Powerful API
- ✅ Email validation features

**Cons:**
- ❌ Requires domain verification
- ❌ Free tier limited to 5,000/month

---

### 5. Amazon SES (Simple Email Service)

**Best For:** AWS environments, high-volume production

**Configuration:**

```json
{
  "smtp_host": "email-smtp.us-east-1.amazonaws.com",
  "smtp_port": 587,
  "username": "your-smtp-username",
  "password": "your-smtp-password",
  "from_address": "verified@yourdomain.com",
  "to_addresses": "alerts@company.com",
  "use_tls": true
}
```

**Setup Steps:**

1. Sign up for AWS account
2. Go to Amazon SES console
3. Verify your email address or domain
4. Create SMTP credentials
5. Request production access (starts in sandbox mode)

**Limits:**
- 200 emails per day (sandbox mode)
- 50,000+ emails per day (production access)
- $0.10 per 1,000 emails (very cheap)

**Pros:**
- ✅ Extremely cheap ($0.10/1000 emails)
- ✅ Unlimited scaling
- ✅ AWS integration
- ✅ High deliverability

**Cons:**
- ❌ Requires AWS account
- ❌ Sandbox mode restrictions initially
- ❌ More complex setup

---

### 6. Postmark

**Best For:** Transactional emails, high deliverability

**Configuration:**

```json
{
  "smtp_host": "smtp.postmarkapp.com",
  "smtp_port": 587,
  "username": "your-server-token",
  "password": "your-server-token",
  "from_address": "alerts@yourdomain.com",
  "to_addresses": "team@company.com",
  "use_tls": true
}
```

**Setup Steps:**

1. Sign up at: https://postmarkapp.com
2. Create a server
3. Get server API token
4. Use token as both username and password
5. Verify sender signature

**Limits:**
- 100 emails per month (free trial)
- 10,000+ emails per month (paid plans from $15/month)

**Pros:**
- ✅ Excellent deliverability
- ✅ Fast delivery
- ✅ Great support
- ✅ Detailed analytics

**Cons:**
- ❌ No free tier (only trial)
- ❌ More expensive than competitors

---

### 7. Local SMTP Server (Self-Hosted)

**Best For:** On-premise deployments, complete control

**Configuration:**

```json
{
  "smtp_host": "mail.company.local",
  "smtp_port": 25,
  "username": "",
  "password": "",
  "from_address": "snapshot-manager@company.local",
  "to_addresses": "it-team@company.local",
  "use_tls": false
}
```

**Common Servers:**
- **Postfix** (Linux)
- **Microsoft Exchange** (Windows)
- **Sendmail** (Unix)
- **Exim** (Linux)

**Pros:**
- ✅ Complete control
- ✅ No external dependencies
- ✅ No sending limits
- ✅ No costs

**Cons:**
- ❌ Requires maintenance
- ❌ Deliverability challenges
- ❌ Security responsibility

---

## 📊 Comparison Table

| Service | Free Tier | Paid Starting | Best For | Setup Difficulty |
|---------|-----------|---------------|----------|------------------|
| Gmail | 500/day | N/A | Testing | ⭐ Easy |
| Microsoft 365 | 300/day | $6/user/month | Enterprise | ⭐ Easy |
| SendGrid | 100/day | $15/month | Production | ⭐⭐ Medium |
| Mailgun | 5,000/month | $35/month | Developers | ⭐⭐ Medium |
| Amazon SES | 200/day* | $0.10/1000 | AWS/Scale | ⭐⭐⭐ Hard |
| Postmark | 100 trial | $15/month | Transactional | ⭐⭐ Medium |
| Self-Hosted | Unlimited | Server costs | On-premise | ⭐⭐⭐⭐ Very Hard |

*Sandbox mode, requires production access request

---

## 🎯 Recommendations by Use Case

### For Testing / Development
**Recommended:** Gmail or Outlook.com
- Easy setup with personal account
- Free tier sufficient for testing
- No credit card required

### For Small Production (< 1,000 emails/day)
**Recommended:** SendGrid or Mailgun
- Reliable delivery
- Good free tiers
- Easy to scale

### For Medium Production (1,000-10,000 emails/day)
**Recommended:** SendGrid or Amazon SES
- Cost-effective
- High reliability
- Good analytics

### For Large Production (> 10,000 emails/day)
**Recommended:** Amazon SES or SendGrid
- Unlimited scaling
- Very low cost per email
- Enterprise features

### For Enterprise / On-Premise
**Recommended:** Microsoft 365 or Self-Hosted
- Integration with existing infrastructure
- Complete control
- Compliance requirements

---

## 🔧 Configuration Examples

### Gmail with App Password

```bash
# In Snapshot Manager UI:
Name: Gmail Alerts
Type: Email
SMTP Host: smtp.gmail.com
SMTP Port: 587
Username: myemail@gmail.com
Password: abcd efgh ijkl mnop  # 16-char app password
From: myemail@gmail.com
To: admin@company.com, ops@company.com
Use TLS: ✓
```

### SendGrid for Production

```bash
# In Snapshot Manager UI:
Name: Production Alerts
Type: Email
SMTP Host: smtp.sendgrid.net
SMTP Port: 587
Username: apikey
Password: SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
From: noreply@yourdomain.com
To: alerts@company.com
Use TLS: ✓
```

### Microsoft 365 for Enterprise

```bash
# In Snapshot Manager UI:
Name: Enterprise Alerts
Type: Email
SMTP Host: smtp.office365.com
SMTP Port: 587
Username: alerts@company.com
Password: your-password-or-app-password
From: alerts@company.com
To: storage-team@company.com
Use TLS: ✓
```

---

## 🐛 Troubleshooting

### "Authentication Failed"

**Gmail:**
- Use app-specific password, not regular password
- Enable 2FA first
- Check "Less secure app access" (not recommended)

**Microsoft 365:**
- Verify SMTP AUTH is enabled
- Use app password if MFA enabled
- Check account not locked

**SendGrid/Mailgun:**
- Verify API key is correct
- Check API key has "Mail Send" permission
- Ensure username is "apikey" for SendGrid

### "Connection Timeout"

- Check firewall allows outbound SMTP (port 587/465)
- Verify SMTP host is correct
- Try alternative port (587 vs 465)
- Check if VPN/proxy blocking

### "Emails Not Received"

- Check spam/junk folder
- Verify recipient address correct
- Check sender domain reputation
- Review service's delivery logs

### "Rate Limit Exceeded"

- Check service's sending limits
- Increase throttling in alert rules
- Upgrade to paid plan
- Use multiple channels

---

## 🔒 Security Best Practices

1. **Use App-Specific Passwords**
   - Never use your main account password
   - Generate unique passwords per application

2. **Enable TLS**
   - Always use TLS/SSL encryption
   - Use port 587 (STARTTLS) or 465 (SSL)

3. **Rotate Credentials**
   - Change passwords regularly
   - Revoke unused API keys

4. **Limit Permissions**
   - Use dedicated email accounts
   - Grant minimum required permissions

5. **Monitor Usage**
   - Track sending volumes
   - Watch for unusual activity
   - Set up alerts for failures

---

## 📈 Scaling Considerations

### When to Upgrade

**From Free to Paid:**
- Hitting daily/monthly limits
- Need better deliverability
- Require analytics/reporting
- Need dedicated IP

**From Basic to Enterprise:**
- Sending > 100,000 emails/month
- Need SLA guarantees
- Require dedicated support
- Compliance requirements

### Multi-Channel Strategy

For high availability, configure multiple channels:

1. **Primary:** SendGrid (high volume)
2. **Backup:** Mailgun (failover)
3. **Critical:** PagerDuty (urgent alerts)

---

## 💡 Tips & Tricks

### Improve Deliverability

1. **Verify Domain:** Use SPF, DKIM, DMARC records
2. **Warm Up:** Gradually increase sending volume
3. **Clean Lists:** Remove bounced addresses
4. **Good Content:** Avoid spam trigger words
5. **Monitor:** Track bounce and complaint rates

### Reduce Costs

1. **Batch Notifications:** Group similar alerts
2. **Smart Throttling:** Reduce duplicate alerts
3. **Use Free Tiers:** Combine multiple services
4. **Optimize Rules:** Only alert on important events

### Testing

1. **Use Test Mode:** Most services have sandbox
2. **Test Addresses:** Use + addressing (user+test@gmail.com)
3. **Monitor Logs:** Check notification history
4. **Verify Delivery:** Confirm emails received

---

## 📚 Additional Resources

### Documentation Links

- **Gmail:** https://support.google.com/mail/answer/7126229
- **Microsoft 365:** https://docs.microsoft.com/en-us/exchange/mail-flow-best-practices/how-to-set-up-a-multifunction-device-or-application-to-send-email-using-microsoft-365-or-office-365
- **SendGrid:** https://docs.sendgrid.com/for-developers/sending-email/getting-started-smtp
- **Mailgun:** https://documentation.mailgun.com/en/latest/user_manual.html#sending-via-smtp
- **Amazon SES:** https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp.html
- **Postmark:** https://postmarkapp.com/developer/user-guide/send-email-with-smtp

---

**Last Updated:** June 11, 2026  
**For:** IBM Storage Virtualize Snapshot Manager

---

* - Your AI Software Engineer*