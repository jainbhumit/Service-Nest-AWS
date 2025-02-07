package model

type OTP struct {
	Email string `json:"pk" dynamodbav:"PK"`
	Otp   string `json:"otp" dynamodbav:"SK"`
	TTl   int64  `json:"ttl" dynamodbav:"ttl"`
}
