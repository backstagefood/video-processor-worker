package utils

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
)

func SendEmail(emailAddress, title, body string) {
	slog.Info("enviando email de erro para usuario", "emailAddress", emailAddress)
	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 587

	// Sender data
	from := os.Getenv("EMAIL_SENDER_FROM")
	password := os.Getenv("EMAIL_SENDER_PASSWORD")

	if len(from) == 0 || len(password) == 0 {
		slog.Info("email ou senha não foram configurados", "email", from, "password", password, "smtpHost", smtpHost, "smtpPort", smtpPort)
		return
	}

	// Receiver email address
	to := []string{
		emailAddress,
	}

	// Message
	formattedMessage := fmt.Sprintf("Subject: %s\r\n\r\n%s", title, body)
	message := []byte(formattedMessage)

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", smtpHost, smtpPort),
		auth,
		from,
		to,
		message,
	)

	if err != nil {
		slog.Error("não foi possível enviar email", "error", err)
	}
}
