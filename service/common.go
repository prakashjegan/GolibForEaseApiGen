package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/libgo/timestring"
	"github.com/prakashjegan/golangexercise/config"
	"github.com/prakashjegan/golangexercise/database"
	"github.com/prakashjegan/golangexercise/database/model"
	influmodel "github.com/prakashjegan/golangexercise/golangexercise-service/database/model"
	"github.com/prakashjegan/golangexercise/golangexercise-service/utils"
	"github.com/prakashjegan/golangexercise/lib"
	"github.com/prakashjegan/golangexercise/lib/middleware"
)

// GetClaims - get JWT custom claims
func GetClaims(c *gin.Context) middleware.MyCustomClaims {
	// get claims
	claims := middleware.MyCustomClaims{
		AuthID:  c.GetUint64("authID"),
		Email:   c.GetString("email"),
		Role:    c.GetString("role"),
		Scope:   c.GetString("scope"),
		TwoFA:   c.GetString("tfa"),
		SiteLan: c.GetString("siteLan"),
		Custom1: c.GetString("custom1"),
		Custom2: c.GetString("custom2"),
	}

	return claims
}

// GetClaims - get JWT custom claims
func GetUserFromContext(c *gin.Context) (*influmodel.User, error) {
	// get claims
	user, ok := c.Get("user")
	if !ok {
		return nil, errors.New("Invalid user type in context")
	}
	usr := (user.(influmodel.User))
	return &usr, nil
}

// ValidateUserID - check whether authID or email is missing
func ValidateUserID(authID uint64, email string, userName string) bool {
	email = strings.TrimSpace(email)
	userName = strings.TrimSpace(userName)
	return authID != 0 && (email != "" || userName != "")
}

// Validate2FA validates user-provided OTP
func Validate2FA(encryptedMessage []byte, issuer string, userInput string) ([]byte, string, error) {
	configSecurity := config.GetConfig().Security
	otpByte, err := lib.ValidateTOTP(encryptedMessage, issuer, userInput)
	// client provided invalid OTP / internal error
	if err != nil {
		// client provided invalid OTP
		if len(otpByte) > 0 {
			return otpByte, configSecurity.TwoFA.Status.Invalid, err
		}

		// internal error
		return []byte{}, "", err
	}

	// validated
	return otpByte, configSecurity.TwoFA.Status.Verified, nil
}

// DelMem2FA - delete secrets from memory
func DelMem2FA(authID uint64) {
	delete(model.InMemorySecret2FA, authID)
}

// SendEmail sends a verification/password recovery email if
// - required by the application
// - an external email service is configured
// - a redis database is configured
func SendEmail(email string, emailType int) (bool, string) {
	// send email if required by the application
	appConfig := config.GetConfig()

	// is external email service activated
	if appConfig.EmailConf.Activate != config.Activated {
		return false, ""
	}

	// is verification/password recovery email required
	doSendEmail := false
	if appConfig.Security.VerifyEmail && emailType == model.EmailTypeVerification {
		doSendEmail = true
	}
	if appConfig.Security.RecoverPass && emailType == model.EmailTypePassRecovery {
		doSendEmail = true
	}
	if !doSendEmail {
		return false, ""
	}

	// is redis database activated
	if appConfig.Database.REDIS.Activate != config.Activated {
		return false, ""
	}

	data := struct {
		key   string
		value string
	}{}
	var keyTTL uint64
	var emailTag string
	var code uint64

	// generate verification/password recovery code
	if emailType == model.EmailTypeVerification {
		code = lib.SecureRandomNumber(appConfig.EmailConf.EmailVerificationCodeLength)
		data.key = model.EmailVerificationKeyPrefix + strconv.FormatUint(code, 10)
		keyTTL = appConfig.EmailConf.EmailVerifyValidityPeriod
		emailTag = appConfig.EmailConf.EmailVerificationTag
	}
	if emailType == model.EmailTypePassRecovery {
		code = lib.SecureRandomNumber(appConfig.EmailConf.PasswordRecoverCodeLength)
		data.key = model.PasswordRecoveryKeyPrefix + strconv.FormatUint(code, 10)
		keyTTL = appConfig.EmailConf.PassRecoverValidityPeriod
		emailTag = appConfig.EmailConf.PasswordRecoverTag
	}
	data.value = utils.EncryptDataWithOutError(email)

	// save in redis with expiry time
	client := *database.GetRedis()
	redisConnTTL := appConfig.Database.REDIS.Conn.ConnTTL

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(redisConnTTL)*time.Second)
	defer cancel()

	// Set key in Redis
	r1 := ""
	if err := client.Do(ctx, radix.FlatCmd(&r1, "SET", data.key, data.value)); err != nil {
		log.WithError(err).Error("error code: 401")
	}
	if r1 != "OK" {
		log.Error("error code: 402")
	}

	// Set expiry time
	r2 := 0
	if err := client.Do(ctx, radix.FlatCmd(&r2, "EXPIRE", data.key, keyTTL)); err != nil {
		log.WithError(err).Error("error code: 403")
	}
	if r2 != 1 {
		log.Error("error code: 404")
	}

	// check which email service
	// for Postmark
	if appConfig.EmailConf.Provider == "postmark" {
		htmlModel := lib.HTMLModel(lib.StrArrHTMLModel(appConfig.EmailConf.HTMLModel))
		htmlModel["secret_code"] = code
		htmlModel["email_validity_period"] = timestring.HourMinuteSecond(keyTTL)

		params := PostmarkParams{}
		params.ServerToken = appConfig.EmailConf.APIToken

		if emailType == model.EmailTypeVerification {
			params.TemplateID = appConfig.EmailConf.EmailVerificationTemplateID
		}

		if emailType == model.EmailTypePassRecovery {
			params.TemplateID = appConfig.EmailConf.PasswordRecoverTemplateID
		}

		params.From = appConfig.EmailConf.AddrFrom
		params.To = email
		params.Tag = emailTag
		params.TrackOpens = appConfig.EmailConf.TrackOpens
		params.TrackLinks = appConfig.EmailConf.TrackLinks
		params.MessageStream = appConfig.EmailConf.DeliveryType
		params.HTMLModel = htmlModel

		// send the email
		res, err := Postmark(params)
		if err != nil {
			log.WithError(err).Error("error code: 405")
		}
		if res.Message != "OK" {
			log.Error(res)
		}
	}

	return true, strconv.FormatUint(code, 10)
}

// SendEmail sends a verification/password recovery email if
// - required by the application
// - an external email service is configured
// - a redis database is configured
func SendMobileOTP(completeMobileNumber string) (bool, string) {
	// send email if required by the application
	appConfig := config.GetConfig()

	// is external email service activated
	if appConfig.EmailConf.Activate != config.Activated {
		return false, ""
	}

	// is redis database activated
	if appConfig.Database.REDIS.Activate != config.Activated {
		return false, ""
	}

	data := struct {
		key   string
		value string
	}{}
	var keyTTL uint64
	//var emailTag string
	var code uint64

	code = lib.SecureRandomNumber(appConfig.EmailConf.EmailVerificationCodeLength)
	data.key = model.MobileVerificationKeyPrefix + strconv.FormatUint(code, 10)
	keyTTL = appConfig.MobileConf.MobileVerifyValidityPeriod
	//emailTag = appConfig.EmailConf.EmailVerificationTag
	data.value = utils.EncryptDataWithOutError(completeMobileNumber)

	// save in redis with expiry time
	client := *database.GetRedis()
	redisConnTTL := appConfig.Database.REDIS.Conn.ConnTTL

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(redisConnTTL)*time.Second)
	defer cancel()

	// Set key in Redis
	r1 := ""
	fmt.Printf("Redis Key : %s , Redis Value : %s", data.key, data.value)
	if err := client.Do(ctx, radix.FlatCmd(&r1, "SET", data.key, data.value)); err != nil {
		log.WithError(err).Error("error code: 401")
	}
	if r1 != "OK" {
		log.Error("error code: 402")
	}

	// Set expiry time
	r2 := 0
	if err := client.Do(ctx, radix.FlatCmd(&r2, "EXPIRE", data.key, keyTTL)); err != nil {
		log.WithError(err).Error("error code: 403")
	}
	if r2 != 1 {
		log.Error("error code: 404")
	}

	//TODO : Verify Mobile Number api via Msg91.

	return true, strconv.FormatUint(code, 10)
}
