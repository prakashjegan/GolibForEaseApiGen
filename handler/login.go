package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	log "github.com/sirupsen/logrus"

	"github.com/dongri/phonenumber"
	"github.com/prakashjegan/golangexercise/config"
	"github.com/prakashjegan/golangexercise/database"
	"github.com/prakashjegan/golangexercise/database/model"
	influModel "github.com/prakashjegan/golangexercise/golangexercise-service/database/model"
	"github.com/prakashjegan/golangexercise/golangexercise-service/utils"
	"github.com/prakashjegan/golangexercise/lib"
	"github.com/prakashjegan/golangexercise/lib/middleware"
	middlewarejwt "github.com/prakashjegan/golangexercise/lib/middleware/middlewarejwt"
	"github.com/prakashjegan/golangexercise/service"
)

// Login handles jobs for controller.Login
func Login(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	var v *model.Auth
	var err error
	if len(payload.UserName) > 0 {
		payload.UserName = strings.TrimSpace(payload.UserName)
		v, err = service.GetUserByUserName(payload.UserName)
		if err != nil {
			httpResponse.Message = "UserName not found"
			httpStatusCode = http.StatusNotFound
			return
		}
	} else if len(payload.Email) > 0 {
		payload.Email = strings.TrimSpace(payload.Email)
		if !lib.ValidateEmail(payload.Email) {
			httpResponse.Message = "wrong email address"
			httpStatusCode = http.StatusBadRequest
			return
		}

		v, err := service.GetUserByEmail(payload.Email)
		if err != nil {
			httpResponse.Message = "email not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		// app settings
		configSecurity := config.GetConfig().Security

		// check whether email verification is required
		if configSecurity.VerifyEmail {
			if v.VerifyEmail != model.EmailVerified {
				httpResponse.Message = "email verification required"
				httpStatusCode = http.StatusUnauthorized
				return
			}
		}
	} else if len(payload.MobileNumber) > 0 {
		number := phonenumber.Parse(payload.MobileNumber, payload.CountryCode)

		if len(number) == 0 {
			httpResponse.Message = "wrong Mobile Number"
			httpStatusCode = http.StatusBadRequest
			return
		}

		v, err := service.GetUserByCompleteMobileNumber(number)
		if err != nil {
			httpResponse.Message = "Mobile Number not found"
			httpStatusCode = http.StatusNotFound
			return
		}
		fmt.Println(v.Password)

	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1011")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// custom claims
	claims := middleware.MyCustomClaims{}
	claims.AuthID = v.AuthID
	claims.LoggedInTime = uint64(time.Now().Unix())
	db := database.GetDB()
	user := influModel.User{}
	if err := db.Where("id_auth = ?", v.AuthID).First(&user).Error; err == nil {
		claims.CompleteMobileNumber = utils.DecryptDataWithOutError(user.CompleteMobileNumber)
		claims.CountryCode = user.CountryCode
		claims.Custom1 = user.AuthType
		claims.MobileNumber = utils.DecryptDataWithOutError(user.MobileNumber)
		claims.Email = utils.DecryptDataWithOutError(user.EmailId)
		claims.Role = user.StakeHolderType
		claims.Scope = user.UserType
		claims.UserID = user.UserID
		claims.UserName = user.UserName
		claims.Custom2 = user.UserPictureLink
	}

	//claims.UserID =
	//claims.Role =
	//claims.Scope
	//claims.TwoFA
	//claims.SiteLan
	//claims.Custom1
	//claims.Custom2

	// when 2FA is enabled for this application (ACTIVATE_2FA=yes)
	// app settings
	configSecurity := config.GetConfig().Security
	if configSecurity.Must2FA == config.Activated {
		twoFA := model.TwoFA{}

		// have the user configured 2FA
		if err := db.Where("id_auth = ?", v.AuthID).First(&twoFA).Error; err == nil {
			// 2FA ON
			if twoFA.Status == configSecurity.TwoFA.Status.On {
				claims.TwoFA = twoFA.Status

				// hash user's pass in sha256
				hashPass := sha256.Sum256([]byte(payload.Password))

				// save the hashed pass in memory for OTP validation step
				data2FA := model.Secret2FA{}
				data2FA.PassSHA = hashPass[:]
				model.InMemorySecret2FA[claims.AuthID] = data2FA
			}
		}
	}

	// issue new tokens
	accessJWT, _, err := middlewarejwt.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1012")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middlewarejwt.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1013")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}

// Refresh handles jobs for controller.Refresh
func Refresh(claims middleware.MyCustomClaims) (httpResponse model.HTTPResponse, httpStatusCode int) {

	// check validity
	ok := service.ValidateUserID(claims.AuthID, claims.Email, claims.UserName)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// issue new tokens
	accessJWT, _, err := middlewarejwt.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1021")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middlewarejwt.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1022")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}
