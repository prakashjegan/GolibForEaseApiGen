// Package model contains all the models required
// for a functional database management system
package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/prakashjegan/golangexercise/config"
	utils "github.com/prakashjegan/golangexercise/golangexercise-service/utils"
	"github.com/prakashjegan/golangexercise/lib"
)

// Email verification statuses
const (
	EmailNotVerified       int8 = -1
	EmailVerifyNotRequired int8 = 0
	EmailVerified          int8 = 1
)

const (
	MobileNotVerified       int8 = -1
	MobileVerifyNotRequired int8 = 0
	MobileVerified          int8 = 1
)

// Email type
const (
	EmailTypeVerification int = 1
	EmailTypePassRecovery int = 2
	MobileVerification    int = 3
)

// Redis key prefixes
const (
	EmailVerificationKeyPrefix  string = "golangexercise-email-verification-"
	PasswordRecoveryKeyPrefix   string = "golangexercise-pass-recover-"
	MobileVerificationKeyPrefix string = "golangexercise-mobile-verification-"
)

// Auth model - `auths` table
type Auth struct {
	AuthID               uint64         `gorm:"primaryKey" json:"authID,omitempty"`
	CreatedAt            time.Time      `gorm:"autoCreateTime;index" json:"createdAt,omitempty"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime;index" json:"updatedAt,omitempty"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
	Email                string         `json:"email,omitempty"`
	UserName             string         `json:"username,omitempty"`
	Password             string         `json:"password,omitempty"`
	PreferredLanguage    string         `json:"preferredLanguage,omitempty"`
	VerifyEmail          int8           `json:"-"`
	CountryCode          string         `json:"countryCode,omitempty"`
	MobileNumber         string         `gorm:"index" json:"mobileNumber,omitempty"`
	CompleteMobileNumber string         `gorm:"index" json:"completeMobileNumber,omitempty"`
	OTP                  int16          `json:"otp,omitempty"`
	VerifyMobileNumber   int8           `json:"-"`
}

// UnmarshalJSON ...
func (v *Auth) UnmarshalJSON(b []byte) error {
	aux := struct {
		AuthID               uint64 `json:"authID"`
		Email                string `json:"email,omitempty"`
		UserName             string `json:"username,omitempty"`
		Password             string `json:"password,omitempty"`
		MobileNumber         string `gorm:"index" json:"mobileNumber,omitempty"`
		CountryCode          string `json:"countryCode,omitempty"`
		OTP                  int16  `json:"otp,omitempty"`
		CompleteMobileNumber string `gorm:"index" json:"completeMobileNumber,omitempty"`
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	configSecurity := config.GetConfig().Security

	// check password length
	// if more checks are required i.e. password pattern,
	// add all conditions here
	if len(aux.Password) < configSecurity.UserPassMinLength {
		return errors.New("short password")
	}

	v.AuthID = aux.AuthID
	v.Email, _ = utils.EncryptData(strings.TrimSpace(aux.Email))
	v.UserName = strings.TrimSpace(aux.UserName)
	v.MobileNumber, _ = utils.EncryptData(strings.TrimSpace(aux.MobileNumber))
	v.CompleteMobileNumber, _ = utils.EncryptData(strings.TrimSpace(aux.CompleteMobileNumber))

	config := lib.HashPassConfig{
		Memory:      configSecurity.HashPass.Memory,
		Iterations:  configSecurity.HashPass.Iterations,
		Parallelism: configSecurity.HashPass.Parallelism,
		SaltLength:  configSecurity.HashPass.SaltLength,
		KeyLength:   configSecurity.HashPass.KeyLength,
	}
	pass, err := lib.HashPass(config, aux.Password)
	if err != nil {
		return err
	}
	v.Password = pass

	return nil
}

// MarshalJSON ...
func (v Auth) MarshalJSON() ([]byte, error) {
	aux := struct {
		AuthID               uint64 `json:"authID"`
		Email                string `json:"email"`
		UserName             string `json:"username,omitempty"`
		MobileNumber         string `gorm:"index" json:"mobileNumber,omitempty"`
		CompleteMobileNumber string `gorm:"index" json:"completeMobileNumber,omitempty"`
	}{
		AuthID:               v.AuthID,
		Email:                strings.TrimSpace(v.Email),
		UserName:             strings.TrimSpace(v.UserName),
		MobileNumber:         strings.TrimSpace(v.MobileNumber),
		CompleteMobileNumber: strings.TrimSpace(v.CompleteMobileNumber),
	}

	return json.Marshal(aux)
}

// AuthPayload - struct to handle all auth data
type AuthPayload struct {
	AuthID               uint64 `json:"authID"`
	Email                string `json:"email,omitempty"`
	Password             string `json:"password,omitempty"`
	MobileNumber         string `gorm:"index" json:"mobileNumber,omitempty"`
	CountryCode          string `json:"countryCode,omitempty"`
	CompleteMobileNumber string `gorm:"index" json:"completeMobileNumber,omitempty"`
	UserName             string `json:"username,omitempty"`
	VerficationType      string `json:"verficationType,omitempty"`

	VerificationCode string `json:"verificationCode,omitempty"`

	OTP string `json:"otp,omitempty"`

	SecretCode  string `json:"secretCode,omitempty"`
	RecoveryKey string `json:"recoveryKey,omitempty"`

	PassNew    string `json:"passNew,omitempty"`
	PassRepeat string `json:"passRepeat,omitempty"`
}
