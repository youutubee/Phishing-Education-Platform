package main

import (
	"fmt"
	"log"
	"os"

	"seap/internal/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Check if API key is set
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ERROR: RESEND_API_KEY is not set in environment variables")
		fmt.Println("\nPlease add RESEND_API_KEY to your .env file:")
		fmt.Println("RESEND_API_KEY=re_your_api_key_here")
		os.Exit(1)
	}

	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
		fmt.Printf("⚠️  WARNING: RESEND_FROM_EMAIL not set, using default: %s\n", fromEmail)
	} else {
		fmt.Printf("✓ Using FROM email: %s\n", fromEmail)
	}

	// Get test email from command line or use default
	testEmail := os.Getenv("TEST_EMAIL")
	if testEmail == "" {
		if len(os.Args) > 1 {
			testEmail = os.Args[1]
		} else {
			fmt.Println("\n❌ ERROR: No test email provided")
			fmt.Println("\nUsage:")
			fmt.Println("  go run test_resend.go your-email@example.com")
			fmt.Println("\nOr set TEST_EMAIL environment variable:")
			fmt.Println("  TEST_EMAIL=your-email@example.com go run test_resend.go")
			os.Exit(1)
		}
	}

	fmt.Printf("✓ Testing Resend API with email: %s\n", testEmail)
	fmt.Println("\nSending test email...")

	// Test the email sending
	subject := "SEAP - Resend API Test"
	html := `
		<div style="font-family: Arial, sans-serif; font-size: 14px; color: #111; padding: 20px;">
			<h2 style="color: #2563eb;">Resend API Test</h2>
			<p>This is a test email from the SEAP (Social Engineering Awareness Platform) system.</p>
			<p>If you received this email, it means your Resend API integration is working correctly! ✅</p>
			<hr style="margin: 20px 0; border: none; border-top: 1px solid #e5e7eb;">
			<p style="color: #6b7280; font-size: 12px;">
				This is an automated test email. You can safely ignore it.
			</p>
		</div>
	`

	err := utils.SendEmail(testEmail, subject, html)

	if err != nil {
		fmt.Printf("\n❌ FAILED: Error sending email\n")
		fmt.Printf("Error details: %v\n", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Verify your RESEND_API_KEY is correct")
		fmt.Println("2. Check that your API key has permission to send emails")
		fmt.Println("3. Ensure the FROM email is verified in your Resend account")
		fmt.Println("4. Check your Resend account dashboard for any errors")
		os.Exit(1)
	}

	fmt.Println("\n✅ SUCCESS: Test email sent successfully!")
	fmt.Printf("Please check the inbox of: %s\n", testEmail)
	fmt.Println("\nNote: It may take a few seconds for the email to arrive.")
}
