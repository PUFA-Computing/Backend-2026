package services

import (
	"Backend/configs"
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sendinblue/APIv3-go-library/v2/lib"
)

// BrevoService is a service for sending emails using Brevo
type BrevoService struct {
	apiKey      string
	senderEmail string
	senderName  string
	apiClient   *lib.APIClient
}

// NewBrevoService creates a new BrevoService
func NewBrevoService(apiKey, senderEmail, senderName string) *BrevoService {
	// Configure API client
	cfg := lib.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	apiClient := lib.NewAPIClient(cfg)

	return &BrevoService{
		apiKey:      apiKey,
		senderEmail: senderEmail,
		senderName:  senderName,
		apiClient:   apiClient,
	}
}

// SendOTPEmail sends an OTP code to the specified email
func (bs *BrevoService) SendOTPEmail(to, otpCode string) error {
	subject := "One Time Password"

	// Use our own HTML template with proper formatting
	body := bs.generateOTPEmailHTML(otpCode)

	return bs.sendEmail(to, subject, body)
}

// SendVerificationEmail sends a verification email with a link
func (bs *BrevoService) SendVerificationEmail(to, token string, userId uuid.UUID) error {
	subject := "Email Verification"

	// Generate verification link using baseURL from config
	baseURL := configs.LoadConfig().BaseURL
	verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s&userId=%s", baseURL, token, userId.String())
	log.Printf("Generated verification link: %s", verificationLink)

	// Generate HTML body
	body := generateVerificationEmailHTML(verificationLink)

	return bs.sendEmail(to, subject, body)
}

// sendEmail sends an email using Brevo API
func (bs *BrevoService) sendEmail(toEmail, subject, htmlContent string) error {
	log.Printf("Attempting to send email to: %s with subject: %s", toEmail, subject)

	// Create sender
	sender := lib.SendSmtpEmailSender{
		Name:  bs.senderName,
		Email: bs.senderEmail,
	}

	// Create recipient
	toList := []lib.SendSmtpEmailTo{
		{
			Email: toEmail,
		},
	}

	// Create email request
	emailRequest := lib.SendSmtpEmail{
		Sender:      &sender,
		To:          toList,
		Subject:     subject,
		HtmlContent: htmlContent,
	}

	// Send the email
	emailsApi := bs.apiClient.TransactionalEmailsApi
	result, _, err := emailsApi.SendTransacEmail(context.Background(), emailRequest)

	if err != nil {
		log.Printf("Error sending email via Brevo: %v", err)
		return fmt.Errorf("failed to send email via Brevo: %w", err)
	}

	log.Printf("Brevo email sent successfully with message ID: %s", result.MessageId)
	return nil
}

// generateOTPEmailHTML creates HTML content for OTP emails with proper formatting
func (bs *BrevoService) generateOTPEmailHTML(otpCode string) string {
	// You can reuse the same HTML template from your SendGrid service
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your OTP Code</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="text-align: center; margin-bottom: 20px;">
        <img src="https://sg.pufacomputing.live/Logo%%20Puma.png" alt="PUFA Computing Logo" width="150" style="max-width: 100%%;">
    </div>
    <div style="background-color: #f9f9f9; border-radius: 5px; padding: 20px; border-top: 3px solid #003CE5;">
        <h1 style="color: #000; text-align: center; margin-bottom: 20px;">Your OTP Code</h1>
        <p style="text-align: center; font-size: 16px; color: #666;">Use the following code to verify your identity:</p>
        <div style="background-color: #eee; padding: 15px; text-align: center; border-radius: 5px; margin: 20px 0; font-size: 24px; letter-spacing: 5px; font-weight: bold;">
            %s
        </div>
        <p style="text-align: center; font-size: 14px; color: #888;">This code will expire in 10 minutes.</p>
    </div>
    <div style="text-align: center; margin-top: 20px; font-size: 12px; color: #999;">
        <p> 2025 PUFA Computing. All rights reserved.</p>
        <p><a href="https://compsci.president.ac.id" style="color: #003CE5; text-decoration: none;">compsci.president.ac.id</a></p>
    </div>
</body>
</html>
`, otpCode)
}
