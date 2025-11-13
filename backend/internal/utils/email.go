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
// Uses the actual email content written by the user when creating the campaign
func SendCampaignShareEmail(to, campaignTitle, emailContent, simulationLink string) error {
	// Use campaign title as subject, or default
	subject := campaignTitle
	if subject == "" {
		subject = "Important: Action Required"
	}

	var builder strings.Builder
	builder.WriteString("<div style=\"font-family: Arial, sans-serif; font-size: 14px; color: #111; line-height: 1.6;\">")
	builder.WriteString("<div style=\"padding: 20px;\">")

	// Convert newlines to HTML line breaks and preserve the user's email content
	// Replace newlines with <br> tags
	formattedContent := strings.ReplaceAll(emailContent, "\n", "<br>")

	// If the content contains the simulation link placeholder or needs the link added
	// Replace any placeholder with the actual link, or append it if not present
	if strings.Contains(formattedContent, "[LINK]") || strings.Contains(formattedContent, "{link}") || strings.Contains(formattedContent, "{{link}}") {
		formattedContent = strings.ReplaceAll(formattedContent, "[LINK]", simulationLink)
		formattedContent = strings.ReplaceAll(formattedContent, "{link}", simulationLink)
		formattedContent = strings.ReplaceAll(formattedContent, "{{link}}", simulationLink)
	} else {
		// If no link placeholder found, add the link at the end as a clickable button
		formattedContent += fmt.Sprintf("<br><br><div style=\"text-align: center; margin: 30px 0;\"><a href=\"%s\" style=\"background-color: #dc2626; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold; font-size: 16px;\">Click Here</a></div>", simulationLink)
	}

	// Add the user's email content
	builder.WriteString(formattedContent)

	builder.WriteString("</div>")
	builder.WriteString("</div>")

	return sendEmail(to, subject, builder.String())
}
