package util

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net/smtp"
	"os"
)

func GenerateUniqueID() string {
	return fmt.Sprintf("%d", rand.Intn(10000))
}

func GenerateUUID() string {
	return uuid.New().String()
}

func sendEmail(to, subject, body string) error {
	from := "bhumitjain9636@gmail.com"
	appPassword := os.Getenv("APP_PASSWORD")

	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message body
	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	// Authentication
	auth := smtp.PlainAuth("", from, appPassword, smtpHost)

	// Sending email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return err
	}

	fmt.Println("Email sent successfully")
	return nil
}

func SendOTPEmail(to, otp string) error {
	subject := "Your OTP Code For Service Nest"
	body := fmt.Sprintf("Your OTP for verification is: %s. It is valid for 5 minutes.", otp)
	return sendEmail(to, subject, body)
}
