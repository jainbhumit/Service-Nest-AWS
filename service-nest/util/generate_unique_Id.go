package util

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"math/rand"
	"net/smtp"
	"service-nest/config"
)

func GenerateUniqueID() string {
	return fmt.Sprintf("%d", rand.Intn(10000))
}

func GenerateUUID() string {
	return uuid.New().String()
}

func sendEmail(to, subject, body string) error {
	cfg := config.Current()
	if cfg == nil {
		return errors.New("application config is not loaded")
	}
	if !cfg.SMTPConfigured() {
		return errors.New("smtp is not configured")
	}

	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	auth := smtp.PlainAuth("", cfg.SMTPFrom, cfg.SMTPAppPassword, cfg.SMTPHost)
	addr := cfg.SMTPHost + ":" + cfg.SMTPPort

	if err := smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{to}, message); err != nil {
		return err
	}

	return nil
}

func SendOTPEmail(to, otp string) error {
	subject := "Your OTP Code For Service Nest"
	body := fmt.Sprintf("Your OTP for verification is: %s. It is valid for 5 minutes.", otp)
	return sendEmail(to, subject, body)
}

func GenerateExclusiveStartKey(pk, sk string) (map[string]types.AttributeValue, error) {
	exclusiveStartKey := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	return exclusiveStartKey, nil
}
