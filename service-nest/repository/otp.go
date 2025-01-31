package repository

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// In-memory OTP store (for demonstration)
var otpStore = struct {
	sync.RWMutex
	data   map[string]string
	expiry map[string]time.Time
}{data: make(map[string]string), expiry: make(map[string]time.Time)}

type OtpRepository struct {
}

func NewOtpRepository() *OtpRepository {
	return &OtpRepository{}
}
func (o *OtpRepository) GenerateOTP() (string, error) {
	otp := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", num)
	}
	return otp, nil
}

func (o *OtpRepository) SaveOTP(email, otp string) {
	otpStore.Lock()
	defer otpStore.Unlock()
	otpStore.data[email] = otp
	otpStore.expiry[email] = time.Now().Add(5 * time.Minute)
}

func (o *OtpRepository) ValidateOTP(email, otp string) bool {
	otpStore.RLock()
	defer otpStore.RUnlock()
	savedOtp, exists := otpStore.data[email]
	if !exists || savedOtp != otp || time.Now().After(otpStore.expiry[email]) {
		return false
	}
	delete(otpStore.data, email)
	delete(otpStore.expiry, email)
	return true
}
