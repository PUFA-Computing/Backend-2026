package configs

import (
	"fmt"
	"os"
)

func getEnvFallback(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisURL  string
	RedisPass string

	ApiPort      string
	JWTSecretKey string

	CloudflareAccountId   string
	CloudflareR2AccessId  string
	CloudflareR2AccessKey string

	S3Endpoint         string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3UsePathStyle     bool
	S3Bucket           string

	// Email service toggle
	UseSmtp bool

	// Legacy SMTP settings
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SenderEmail  string

	// SendGrid settings
	SendGridAPIKey     string
	SendGridSender     string
	SendGridSenderName string

	// Brevo settings
	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string

	GithubAccessToken string
	HunterApiKey      string

	BaseURL string
}

func LoadConfig() *Config {
	env := os.Getenv("ENV")

	var baseURl string

	if env == "production" {
		baseURl = "https://compsci.president.ac.id"
	} else if env == "staging" {
		baseURl = "https://staging.compsci.president.ac.id"
	} else if env == "local" || env == "test" {
		baseURl = "http://localhost:3000"
	} else {
		// Default to production URL if ENV is not explicitly set to local/test
		baseURl = "https://compsci.president.ac.id"
	}

	cfg := &Config{
		DBHost:                os.Getenv("DB_HOST"),
		DBPort:                os.Getenv("DB_PORT"),
		DBUser:                os.Getenv("DB_USER"),
		DBPassword:            os.Getenv("DB_PASSWORD"),
		DBName:                os.Getenv("DB_NAME"),
		RedisURL:              os.Getenv("REDIS_URL"),
		RedisPass:             os.Getenv("REDIS_PASS"),
		ApiPort:               os.Getenv("API_PORT"),
		JWTSecretKey:          os.Getenv("JWT_SECRET_KEY"),
		CloudflareAccountId:   os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		CloudflareR2AccessId:  os.Getenv("CLOUDFLARE_R2_ACCESS_ID"),
		CloudflareR2AccessKey: os.Getenv("CLOUDFLARE_R2_ACCESS_KEY"),
		S3Endpoint:            os.Getenv("S3_ENDPOINT"),
		S3UsePathStyle:        os.Getenv("S3_USE_PATH_STYLE") == "true",
		AWSAccessKeyId:        os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:             os.Getenv("AWS_REGION"),
		S3Bucket:              os.Getenv("S3_BUCKET"),
		// Email service toggle (FORCED TRUE for SMTP instead of Brevo)
		UseSmtp: true,
		// Legacy SMTP settings with hardcoded fallbacks
		SMTPHost:     getEnvFallback("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvFallback("SMTP_PORT", "587"),
		SMTPUsername: getEnvFallback("SMTP_USERNAME", "rnt.compsci@gmail.com"),
		SMTPPassword: getEnvFallback("SMTP_PASSWORD", "wvxayloupcrmqmbg"),
		SenderEmail:  getEnvFallback("SMTP_SENDER_EMAIL", "rnt.compsci@gmail.com"),
		// SendGrid settings
		SendGridAPIKey:     os.Getenv("SENDGRID_API_KEY"),
		SendGridSender:     os.Getenv("SENDGRID_SENDER_EMAIL"),
		SendGridSenderName: os.Getenv("SENDGRID_SENDER_NAME"),
		// Brevo settings
		BrevoAPIKey:       os.Getenv("BREVO_API_KEY"),
		BrevoSenderEmail:  os.Getenv("BREVO_SENDER_EMAIL"),
		BrevoSenderName:   os.Getenv("BREVO_SENDER_NAME"),
		BaseURL:           baseURl,
		GithubAccessToken: os.Getenv("GH_ACCESS_TOKEN"),
		HunterApiKey:      os.Getenv("HUNTER_API_KEY"),
	}

	fmt.Printf("Loaded Config: %+v\n", cfg)
	return cfg
}
