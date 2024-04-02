package config

import (
	"crypto"

	"github.com/prakashjegan/golangexercise/lib"
	"github.com/prakashjegan/golangexercise/lib/middleware"
)

// SecurityConfig ...
type SecurityConfig struct {
	UserPassMinLength int

	MustBasicAuth                string
	PERSONAL_DATA_ENCRYPTION_KEY string
	BasicAuth                    struct {
		Username string
		Password string
	}

	GPT_API_KEY string
	MustJWT     string
	JWT         middleware.JWTParameters

	MustHash string
	HashPass lib.HashPassConfig

	VerifyEmail bool
	RecoverPass bool

	MustFW   string
	Firewall struct {
		ListType string
		IP       string
	}

	MustCORS string
	CORS     []middleware.CORSPolicy

	TrustedPlatform string

	Must2FA string
	TwoFA   struct {
		Issuer string
		Crypto crypto.Hash
		Digits int

		Status Status2FA
		PathQR string
	}
}

// Status2FA - user's 2FA statuses
type Status2FA struct {
	Verified string
	On       string
	Off      string
	Invalid  string
}

// AwsConfig - AwsConfig statuses
type AwsConfig struct {
	Activate              string
	AccessKey             string
	SecreteAccessKey      string
	Region                string
	DocumentBucketName    string
	InvoiceBucketName     string
	PublicImageBucketName string
}

type GoogleConfig struct {
	Activate               string
	AccessKey              string
	SecreteAccessKey       string
	Region                 string
	GoogleSheetId          string
	ContentConversionSheet string
	PlatformMetricSheet    string
	Path                   string
}

type FireBaseConfig struct {
	Activate               string
	AccessKey              string
	SecreteAccessKey       string
	Region                 string
	GoogleSheetId          string
	ContentConversionSheet string
	PlatformMetricSheet    string
	Path                   string

	OneSignalAPPId  string
	OneSignalAPIKey string
}
