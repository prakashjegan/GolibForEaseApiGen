package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dongri/phonenumber"
	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	"github.com/prakashjegan/golangexercise/config"
	"github.com/prakashjegan/golangexercise/database"
	"github.com/prakashjegan/golangexercise/database/model"
	influModel "github.com/prakashjegan/golangexercise/golangexercise-service/database/model"
	serv "github.com/prakashjegan/golangexercise/golangexercise-service/service"
	"github.com/prakashjegan/golangexercise/golangexercise-service/utils"
	"github.com/prakashjegan/golangexercise/lib"
	"github.com/prakashjegan/golangexercise/service"
	"github.com/prakashjegan/golangexercise/uid64"
)

// VerifyEmail handles jobs for controller.VerifyEmail
func VerifyEmail(payload model.AuthPayload, user *influModel.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	data := struct {
		key   string
		value string
	}{}
	data.key = model.EmailVerificationKeyPrefix + payload.VerificationCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1061")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Message = "wrong/expired verification code"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1062")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1063")
	}
	if result == 0 {
		err := errors.New("failed to delete recovery key from redis")
		log.WithError(err).Error("error code: 1064")
	}

	// update verification status in database
	db := database.GetDB()
	auth := model.Auth{}

	if err := db.Where("auth_id = ?", user.IDAuth).First(&auth).Error; err != nil {
		httpResponse.Message = "unknown user"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	verifi := influModel.VerificationCode{}
	if err := db.Where(" verification_string = ? and user_id = ?  and verification_type = ? ", data.value, user.UserID, payload.VerficationType).First(&verifi).Error; err == nil {

	}
	canUpdateAuth := false
	if verifi.VerificationType == "BASE_EMAIL" {
		canUpdateAuth = true
	}

	tx := db.Begin()
	if canUpdateAuth {
		if auth.Email != data.value {
			auth.Email = data.value
		}
		if auth.VerifyEmail == model.EmailVerified {
			httpResponse.Message = "email already verified"
			httpStatusCode = http.StatusOK
			return
		}

		auth.VerifyEmail = model.EmailVerified
		auth.UpdatedAt = time.Now().Local()

		if err := tx.Save(&auth).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1065")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		if err := db.Model(&influModel.User{}).Where("id_auth = ?", auth.AuthID).Updates(map[string]interface{}{"verify_email": model.EmailVerified, "email_id": data.value}).Error; err != nil {
			log.WithError(err).Error("error code: 1111")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		if err := db.Model(&influModel.Partner{}).Where("id_auth = ?", auth.AuthID).Updates(map[string]interface{}{"verify_email": model.EmailVerified, "email_id": data.value}).Error; err != nil {
			log.WithError(err).Error("error code: 1111")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	uid1, _ := uid64.New()
	verificationCode := influModel.VerificationCode{}
	verificationCode.VerificationCodeId = uint64(uid1)
	verificationCode.OTP = payload.VerificationCode
	verificationCode.VerificationString = verifi.VerificationStatus
	verificationCode.VerificationType = payload.VerficationType
	verificationCode.VerificationStatus = "VERIFIED"
	verificationCode.UserId = user.UserID
	verificationCode.IdAuth = user.IDAuth

	verificationC := influModel.VerificationCode{}
	switch verificationCode.VerificationType {
	case "BASE_EMAIL":
		if err := db.Where(" verification_string = ? and user_id = ?  ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	case "BASE_MOBILE":
		if err := db.Where(" verification_string = ?  and user_id = ?   ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	}

	if err := tx.Save(&verificationCode).Error; err != nil {
		tx.Rollback()
		httpResponse.Message = "sent verification email failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	tx.Commit()

	httpResponse.Message = "email successfully verified"
	httpStatusCode = http.StatusOK
	return
}

// CreateVerificationEmail handles jobs for controller.CreateVerificationEmail
func CreateVerificationEmail(payload model.AuthPayload, user *influModel.User) (httpResponse model.HTTPResponse, httpStatusCode int) {

	db := database.GetDB()
	//user := influModel.User{}
	// if err := db.Where("id_auth = ?", payload.AuthID).First(&user).Error; err == nil {
	// 	log.WithError(err).Error("error code: 1111")
	// 	httpResponse.Message = "internal server error"
	// 	httpStatusCode = http.StatusInternalServerError
	// 	return
	// }

	//payload.Email = utils.DecryptDataWithOutError(strings.TrimSpace(user.EmailId))
	payload.Email = strings.TrimSpace(user.EmailId)
	if !lib.ValidateEmail(payload.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	//v, err := service.GetUserByEmail(strings.TrimSpace(user.EmailId))
	v, err := service.GetUserById(user.UserID)

	if err != nil {
		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// is email already verified
	verificationC := influModel.VerificationCode{}
	if err := db.Where(" verification_string = ? and user_id = ? ", utils.EncryptDataWithOutError(payload.Email), user.UserID).First(&verificationC).Error; err == nil {
		if verificationC.VerificationStatus == "VERIFIED" {
			httpResponse.Message = "mobile already verified"
			httpStatusCode = http.StatusOK
			return
		}
	}

	// is email already verified
	// if v.VerifyEmail == model.EmailVerified {
	// 	httpResponse.Message = "email already verified"
	// 	httpStatusCode = http.StatusOK
	// 	return
	// }

	// // verify password
	// verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	// if err != nil {
	// 	log.WithError(err).Error("error code: 1071")
	// 	httpResponse.Message = "internal server error"
	// 	httpStatusCode = http.StatusInternalServerError
	// 	return
	// }
	// if !verifyPass {
	// 	httpResponse.Message = "wrong credentials"
	// 	httpStatusCode = http.StatusUnauthorized
	// 	return
	// }

	// issue new verification code
	verifyDone, code := service.SendEmail(payload.Email, model.EmailTypeVerification)
	if !verifyDone {
		httpResponse.Message = "failed to send verification email"
		httpStatusCode = http.StatusServiceUnavailable
		return
	}

	uid1, _ := uid64.New()
	verificationCode := influModel.VerificationCode{}
	verificationCode.VerificationCodeId = uint64(uid1)
	verificationCode.OTP = code
	verificationCode.VerificationString = utils.EncryptDataWithOutError(payload.Email)
	verificationCode.VerificationType = "BASE_EMAIL"
	verificationCode.VerificationStatus = "INITIATED"
	verificationCode.UserId = user.UserID
	verificationCode.IdAuth = user.IDAuth

	switch verificationCode.VerificationType {
	case "BASE_EMAIL":
		if err := db.Where(" verification_string = ? and user_id = ?  ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	case "BASE_MOBILE":
		if err := db.Where(" verification_string = ?  and user_id = ?   ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	}

	tx := db.Begin()
	if err := tx.Save(&verificationCode).Error; err != nil {
		tx.Rollback()
		httpResponse.Message = "sent verification email failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if v.EmailId != utils.EncryptDataWithOutError(payload.Email) {
		v.EmailId = utils.EncryptDataWithOutError(payload.Email)
		v.VerifyEmail = model.EmailNotVerified
		if err := tx.Save(&v).Error; err != nil {
			tx.Rollback()
			httpResponse.Message = "sent verification email failed"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	tx.Commit()

	httpResponse.Message = "sent verification email"
	httpStatusCode = http.StatusOK
	return
}

// VerifyEmail handles jobs for controller.VerifyEmail
func VerifyMobile(payload model.AuthPayload, user *influModel.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	data := struct {
		key   string
		value string
	}{}
	data.key = model.MobileVerificationKeyPrefix + payload.VerificationCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1061")
		httpResponse.Message = "Wrong OTP"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Message = "Wrong/expired verification code"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1062")
		httpResponse.Message = "Wrong OTP"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1063")
	}
	if result == 0 {
		err := errors.New("failed to delete recovery key from redis")
		log.WithError(err).Error("error code: 1064")
	}

	// update verification status in database
	db := database.GetDB()
	auth := model.Auth{}
	// if err := db.Where("complete_mobile_number = ?", data.value).First(&auth).Error; err != nil {
	// 	httpResponse.Message = "unknown user"
	// 	httpStatusCode = http.StatusUnauthorized
	// 	return
	// }
	if err := db.Where("auth_id = ?", user.IDAuth).First(&auth).Error; err != nil {
		httpResponse.Message = "unknown user"
		httpStatusCode = http.StatusUnauthorized
		return
	}
	verifi := influModel.VerificationCode{}
	fmt.Printf("Verification_string : %s , user_id : %s , verification_type : %s ", data.value, user.UserID, payload.VerficationType)
	if err := db.Where(" verification_string = ? and user_id = ?  and verification_type = ? ", data.value, user.UserID, payload.VerficationType).First(&verifi).Error; err == nil {

	}
	canUpdateAuth := false
	if verifi.VerificationType == "BASE_MOBILE" {
		canUpdateAuth = true
	}

	tx := db.Begin()
	if canUpdateAuth {

		if auth.CompleteMobileNumber != data.value {
			// auth.CompleteMobileNumber = data.value
			// auth.MobileNumber = verifi.MobileNumber
			httpResponse.Message = "Invalid OTP"
			httpStatusCode = http.StatusUnauthorized
			return
		}

		if auth.VerifyMobileNumber == model.MobileVerified {
			httpResponse.Message = "Mobile already verified"
			httpStatusCode = http.StatusOK
			return
		}

		auth.VerifyMobileNumber = model.MobileVerified
		auth.UpdatedAt = time.Now().Local()

		if err := tx.Save(&auth).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1065")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		if err := db.Model(&influModel.User{}).Where("id_auth = ?", auth.AuthID).Updates(map[string]interface{}{"verify_mobile_number": model.MobileVerified, "complete_mobile_number": data.value, "mobile_number": verifi.MobileNumber}).Error; err != nil {
			log.WithError(err).Error("error code: 1111")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		if err := db.Model(&influModel.Partner{}).Where("id_auth = ?", auth.AuthID).Updates(map[string]interface{}{"verify_mobile_number": model.MobileVerified, "complete_mobile_number": data.value, "mobile_number": verifi.MobileNumber}).Error; err != nil {
			log.WithError(err).Error("error code: 1111")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	uid1, _ := uid64.New()
	verificationCode := influModel.VerificationCode{}
	verificationCode.VerificationCodeId = uint64(uid1)
	verificationCode.OTP = payload.VerificationCode
	verificationCode.VerificationString = verifi.VerificationString
	verificationCode.VerificationType = payload.VerficationType
	verificationCode.MobileNumber = verifi.MobileNumber
	verificationCode.VerificationStatus = "VERIFIED"
	verificationCode.UserId = user.UserID
	verificationCode.IdAuth = user.IDAuth

	verificationC := influModel.VerificationCode{}
	switch verificationCode.VerificationType {
	case "BASE_EMAIL":
		if err := db.Where(" verification_string = ? and user_id = ?  ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	case "BASE_MOBILE":
		if err := db.Where(" verification_string = ?  and user_id = ?   ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	}

	if err := tx.Save(&verificationCode).Error; err != nil {
		tx.Rollback()
		httpResponse.Message = "otp verification sms failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	tx.Commit()

	httpResponse.Message = "mobile successfully verified"
	httpStatusCode = http.StatusOK
	return
}

// CreateVerificationEmail handles jobs for controller.CreateVerificationEmail
func CreateVerificationMobile(payload model.AuthPayload, user *influModel.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	//Decrypted Mobile number
	db := database.GetDB()

	//v, err := service.GetUserByCompleteMobileNumber(payload.CompleteMobileNumber)
	v, err := service.GetUserById(user.UserID)
	if err != nil {
		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}
	//payload.CompleteMobileNumber = phonenumber.Parse(utils.DecryptDataWithOutError(payload.MobileNumber), v.CountryCode)
	payload.CompleteMobileNumber = phonenumber.Parse(payload.MobileNumber, v.CountryCode)
	payload.CompleteMobileNumber = strings.TrimSpace(payload.CompleteMobileNumber)
	// if !lib.ValidateMobile(utils.DecryptDataWithOutError(payload.Email)) {
	// 	httpResponse.Message = "wrong email address"
	// 	httpStatusCode = http.StatusBadRequest
	// 	return
	// }

	// is email already verified
	verificationC := influModel.VerificationCode{}
	if err := db.Where(" verification_string = ? and user_id = ? ", utils.EncryptDataWithOutError(payload.CompleteMobileNumber), user.UserID).First(&verificationC).Error; err == nil {
		if verificationC.VerificationStatus == "VERIFIED" {
			httpResponse.Message = "mobile already verified"
			httpStatusCode = http.StatusOK
			return
		}
	}

	// // verify password
	// verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	// if err != nil {
	// 	log.WithError(err).Error("error code: 1071")
	// 	httpResponse.Message = "internal server error"
	// 	httpStatusCode = http.StatusInternalServerError
	// 	return
	// }
	// if !verifyPass {
	// 	httpResponse.Message = "wrong credentials"
	// 	httpStatusCode = http.StatusUnauthorized
	// 	return
	// }

	// issue new verification code
	val, code := service.SendMobileOTP(payload.CompleteMobileNumber)
	if !val {
		httpResponse.Message = "failed to send verification email"
		httpStatusCode = http.StatusServiceUnavailable
		return
	}

	uid1, _ := uid64.New()
	verificationCode := influModel.VerificationCode{}
	verificationCode.VerificationCodeId = uint64(uid1)
	verificationCode.OTP = code
	verificationCode.VerificationString = utils.EncryptDataWithOutError(payload.CompleteMobileNumber)
	verificationCode.MobileNumber = utils.EncryptDataWithOutError(payload.MobileNumber)
	verificationCode.VerificationType = "BASE_MOBILE"
	verificationCode.VerificationStatus = "INITIATED"
	verificationCode.UserId = user.UserID
	verificationCode.IdAuth = user.IDAuth

	switch verificationCode.VerificationType {
	case "BASE_EMAIL":
		if err := db.Where(" verification_string = ? and user_id = ?  ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	case "BASE_MOBILE":
		if err := db.Where(" verification_string = ?  and user_id = ?   ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	}

	tx := db.Begin()
	if err := tx.Save(&verificationCode).Error; err != nil {
		tx.Rollback()
		httpResponse.Message = "sent verification sms failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if v.CompleteMobileNumber != utils.EncryptDataWithOutError(payload.CompleteMobileNumber) {
		v.CompleteMobileNumber = utils.EncryptDataWithOutError(payload.CompleteMobileNumber)
		v.MobileNumber = utils.EncryptDataWithOutError(payload.MobileNumber)
		v.VerifyMobileNumber = model.EmailNotVerified
		if err := tx.Save(&v).Error; err != nil {
			tx.Rollback()
			httpResponse.Message = "sent verification sms failed"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	tx.Commit()

	httpResponse.Message = "sent verification sms"
	httpStatusCode = http.StatusOK
	return
}

func CreateVerificationMobileViaSend(payload model.AuthPayload, user *influModel.User, isMobile bool) (httpResponse model.HTTPResponse, httpStatusCode int) {
	//Decrypted Mobile number
	db := database.GetDB()

	//v, err := service.GetUserByCompleteMobileNumber(payload.CompleteMobileNumber)
	v, err := service.GetUserById(user.UserID)
	if err != nil {
		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}
	//payload.CompleteMobileNumber = phonenumber.Parse(utils.DecryptDataWithOutError(payload.MobileNumber), v.CountryCode)
	payload.CompleteMobileNumber = phonenumber.Parse(payload.MobileNumber, v.CountryCode)
	payload.CompleteMobileNumber = strings.TrimSpace(payload.CompleteMobileNumber)
	// if !lib.ValidateMobile(utils.DecryptDataWithOutError(payload.Email)) {
	// 	httpResponse.Message = "wrong email address"
	// 	httpStatusCode = http.StatusBadRequest
	// 	return
	// }

	// is email already verified
	verificationC := influModel.VerificationCode{}
	if err := db.Where(" verification_string = ? and user_id = ? ", utils.EncryptDataWithOutError(payload.CompleteMobileNumber), user.UserID).First(&verificationC).Error; err == nil {
		if verificationC.VerificationStatus == "VERIFIED" {
			httpResponse.Message = "mobile already verified"
			httpStatusCode = http.StatusOK
			return
		}
	}

	// // verify password
	// verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	// if err != nil {
	// 	log.WithError(err).Error("error code: 1071")
	// 	httpResponse.Message = "internal server error"
	// 	httpStatusCode = http.StatusInternalServerError
	// 	return
	// }
	// if !verifyPass {
	// 	httpResponse.Message = "wrong credentials"
	// 	httpStatusCode = http.StatusUnauthorized
	// 	return
	// }

	// issue new verification code
	appConfig := config.GetConfig()
	code := fmt.Sprintf("%d", lib.SecureRandomNumber(appConfig.EmailConf.EmailVerificationCodeLength))
	val := true
	if !val {
		if isMobile {
			httpResponse.Message = "failed to Generate Verification for Mobile"

		} else {
			httpResponse.Message = "failed to Generate Verification for Mobile"
		}

		httpStatusCode = http.StatusServiceUnavailable
		return
	}

	uid1, _ := uid64.New()
	verificationCode := influModel.VerificationCode{}
	verificationCode.VerificationCodeId = uint64(uid1)
	verificationCode.OTP = code
	if isMobile {
		verificationCode.VerificationString = utils.EncryptDataWithOutError(payload.CompleteMobileNumber)
		verificationCode.MobileNumber = utils.EncryptDataWithOutError(payload.MobileNumber)
		verificationCode.VerificationType = "BASE_MOBILE"
	} else {

	}

	verificationCode.VerificationStatus = "INITIATED"
	verificationCode.UserId = user.UserID
	verificationCode.IdAuth = user.IDAuth

	switch verificationCode.VerificationType {
	case "BASE_EMAIL":
		if err := db.Where(" verification_string = ? and user_id = ?  ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	case "BASE_MOBILE":
		if err := db.Where(" verification_string = ?  and user_id = ?   ", verificationCode.VerificationString, user.UserID).First(&verificationC).Error; err == nil {
			verificationCode.VerificationCodeId = verificationC.VerificationCodeId
		} else {
			verificationCode.VerificationCodeId = uint64(uid1)
		}
	}

	tx := db.Begin()
	if err := tx.Save(&verificationCode).Error; err != nil {
		tx.Rollback()
		httpResponse.Message = "sent verification sms failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if v.CompleteMobileNumber != utils.EncryptDataWithOutError(payload.CompleteMobileNumber) {
		v.CompleteMobileNumber = utils.EncryptDataWithOutError(payload.CompleteMobileNumber)
		v.MobileNumber = utils.EncryptDataWithOutError(payload.MobileNumber)
		v.VerifyMobileNumber = model.EmailNotVerified
		if err := tx.Save(&v).Error; err != nil {
			tx.Rollback()
			httpResponse.Message = "sent verification sms failed"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	tx.Commit()
	platfAcc := serv.FetchMobileVerificationOrEmailOTP(isMobile, code)
	httpResponse.Message = platfAcc.VerificationString
	//Verified.

	httpStatusCode = http.StatusOK
	return
}
