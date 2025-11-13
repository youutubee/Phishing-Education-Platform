package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendEmail sends an email using Resend API (public function for testing)
func SendEmail(to, subject, html string) error {
	return sendEmail(to, subject, html)
}

func sendEmail(to, subject, html string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Printf("WARNING: RESEND_API_KEY not set; cannot send email to %s", to)
		return fmt.Errorf("RESEND_API_KEY environment variable is not set")
	}

	from := os.Getenv("RESEND_FROM_EMAIL")
	if from == "" {
		from = "onboarding@resend.dev" // Default Resend test domain
		log.Printf("WARNING: RESEND_FROM_EMAIL not set; using default: %s", from)
	} else {
		// Validate email format - must contain @ symbol
		if !strings.Contains(from, "@") {
			log.Printf("WARNING: RESEND_FROM_EMAIL '%s' is not a valid email format. Using default.", from)
			from = "onboarding@resend.dev"
		}
	}

	payload := resendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal resend payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create resend request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("resend API error: %s - %s", resp.Status, string(respBody))
		log.Printf("Failed to send email to %s: %s", to, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("Successfully sent email to %s with subject: %s", to, subject)
	return nil
}

func SendCampaignDecisionEmail(to, campaignTitle, status, comment, simulationLink string) {
	statusLabel := "Updated"
	if len(status) > 0 {
		statusLabel = strings.ToUpper(status[:1]) + status[1:]
	}
	subject := fmt.Sprintf("Campaign %s: %s", campaignTitle, statusLabel)

	var builder strings.Builder
	builder.WriteString("<div style=\"font-family: Arial, sans-serif; font-size: 14px; color: #111\">")
	builder.WriteString(fmt.Sprintf("<p>Hello,</p><p>Your campaign <strong>%s</strong> has been <strong>%s</strong>.</p>", campaignTitle, statusLabel))

	if comment != "" {
		builder.WriteString(fmt.Sprintf("<p><strong>Admin comment:</strong><br/>%s</p>", comment))
	}

	if status == "approved" && simulationLink != "" {
		builder.WriteString(fmt.Sprintf("<p>You can access the simulation using the link below:</p><p><a href=\"%s\" style=\"color:#2563eb\">%s</a></p>", simulationLink, simulationLink))
	}

	builder.WriteString("<p>Thank you for using SEAP.</p></div>")

	if err := sendEmail(to, subject, builder.String()); err != nil {
		log.Printf("ERROR: Failed to send campaign status email to %s: %v", to, err)
		// Don't return error - email failure shouldn't break the approval/rejection flow
	} else {
		log.Printf("Successfully sent campaign %s notification email to %s", status, to)
	}
}

// SendCampaignShareEmail sends a campaign link to a recipient via email
func SendCampaignShareEmail(to, campaignTitle, simulationLink string) error {
	subject := fmt.Sprintf("Phishing Awareness Campaign: %s", campaignTitle)

	var builder strings.Builder
	builder.WriteString("<div style=\"font-family: Arial, sans-serif; font-size: 14px; color: #111; padding: 20px;\">")
	builder.WriteString(fmt.Sprintf("<h2 style=\"color: #2563eb;\">Phishing Awareness Campaign</h2>"))
	builder.WriteString(fmt.Sprintf("<p>Hello,</p>"))
	builder.WriteString(fmt.Sprintf("<p>You have been invited to participate in a <strong>simulated phishing awareness campaign</strong>: <strong>%s</strong>.</p>", campaignTitle))
	builder.WriteString("<p>This is a safe, educational simulation designed to help you recognize phishing attempts and protect yourself from real threats.</p>")
	builder.WriteString("<p><strong>Click the link below to access the simulation:</strong></p>")
	builder.WriteString(fmt.Sprintf("<p style=\"margin: 20px 0;\"><a href=\"%s\" style=\"background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block;\">Access Campaign</a></p>", simulationLink))
	builder.WriteString(fmt.Sprintf("<p style=\"color: #6b7280; font-size: 12px; margin-top: 20px;\">Or copy this link: <br/><code style=\"background: #f3f4f6; padding: 4px 8px; border-radius: 4px;\">%s</code></p>", simulationLink))
	builder.WriteString("<hr style=\"margin: 20px 0; border: none; border-top: 1px solid #e5e7eb;\">")
	builder.WriteString("<p style=\"color: #6b7280; font-size: 12px;\"><strong>Note:</strong> This is a simulated phishing attempt for educational purposes. No real data will be collected or stored.</p>")
	builder.WriteString("<p style=\"color: #6b7280; font-size: 12px;\">Thank you for participating in this security awareness training.</p>")
	builder.WriteString("</div>")

	return sendEmail(to, subject, builder.String())
}
