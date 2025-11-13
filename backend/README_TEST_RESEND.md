# Resend API Test Script

This script helps you test if your Resend API integration is working correctly.

## Prerequisites

1. Make sure you have a Resend account: https://resend.com
2. Get your API key from the Resend dashboard
3. Add it to your `.env` file in the `backend` directory

## Setup

1. Add your Resend API key to `backend/.env`:
   ```env
   RESEND_API_KEY=re_your_api_key_here
   RESEND_FROM_EMAIL=onboarding@resend.dev
   ```

   Note: For testing, you can use `onboarding@resend.dev` as the FROM email. For production, you'll need to verify your own domain.

## Running the Test

### Option 1: Pass email as command line argument
```bash
cd backend
go run test_resend.go your-email@example.com
```

### Option 2: Use environment variable
```bash
cd backend
TEST_EMAIL=your-email@example.com go run test_resend.go
```

## Expected Output

### Success:
```
✓ Using FROM email: onboarding@resend.dev
✓ Testing Resend API with email: your-email@example.com

Sending test email...

✅ SUCCESS: Test email sent successfully!
Please check the inbox of: your-email@example.com

Note: It may take a few seconds for the email to arrive.
```

### Failure (Missing API Key):
```
❌ ERROR: RESEND_API_KEY is not set in environment variables

Please add RESEND_API_KEY to your .env file:
RESEND_API_KEY=re_your_api_key_here
```

### Failure (API Error):
```
❌ FAILED: Error sending email
Error details: resend API error: 401 - {"message":"Invalid API key"}

Troubleshooting:
1. Verify your RESEND_API_KEY is correct
2. Check that your API key has permission to send emails
3. Ensure the FROM email is verified in your Resend account
4. Check your Resend account dashboard for any errors
```

## Troubleshooting

1. **"RESEND_API_KEY not set"**
   - Make sure you have a `.env` file in the `backend` directory
   - Verify the key name is exactly `RESEND_API_KEY`
   - Restart your terminal/IDE after adding the variable

2. **"Invalid API key"**
   - Double-check your API key from the Resend dashboard
   - Make sure there are no extra spaces or quotes
   - Regenerate the key if needed

3. **"Domain not verified"**
   - For testing, use `onboarding@resend.dev` as FROM email
   - For production, verify your domain in Resend dashboard

4. **Email not received**
   - Check spam/junk folder
   - Wait a few seconds (emails can take time)
   - Verify the recipient email address is correct
   - Check Resend dashboard for delivery status

## Next Steps

Once the test succeeds, your Resend integration is ready! The system will automatically send emails when:
- Campaigns are approved
- Campaigns are rejected

Make sure to check your backend logs for email sending status.

