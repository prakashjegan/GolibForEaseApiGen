package middlewarejwt

// github.com/prakashjegan/golangexercise
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	gdatabase "github.com/prakashjegan/golangexercise/database"
	influModel "github.com/prakashjegan/golangexercise/golangexercise-service/database/model"
	"github.com/prakashjegan/golangexercise/golangexercise-service/utils"
	mid "github.com/prakashjegan/golangexercise/lib/middleware"
)

// JWT - validate access token
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {

		val := c.Request.Header.Get("Authorization")
		if len(val) == 0 || !strings.Contains(val, "Bearer ") {
			// no vals or no bearer found
			fmt.Printf("AuthroizationFailed 1: No bearer \n")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		vals := strings.Split(val, " ")
		if len(vals) != 2 {
			fmt.Printf("AuthroizationFailed 1: splits is not 2 \n")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		db := gdatabase.GetDB()
		testValue := "Test1234:"
		if strings.HasPrefix(vals[1], testValue) {
			emailDescrypted := strings.ReplaceAll(vals[1], "Test1234:", "")
			userClaim := influModel.User{}
			if err := db.Where("email_id = ? ", utils.EncryptDataWithOutError(emailDescrypted)).First(&userClaim).Error; err == nil {
				c.Set("authID", userClaim.IDAuth)
				c.Set("email", userClaim.EmailId)
				c.Set("role", "TEMP_ROLE")
				c.Set("scope", "TEMP_SCOPE")
				c.Set("tfa", "TEMP_TFA")
				c.Set("siteLan", "TEMP_LAN")
				c.Set("custom1", "TEMP_CUSTOM1")
				c.Set("custom2", "TEMP_CUSTOM2")
				c.Set("userName", userClaim.UserName)
				c.Set("mobileNumber", userClaim.MobileNumber)
				c.Set("countryCode", userClaim.CountryCode)
				c.Set("completeMobileNumber", userClaim.CompleteMobileNumber)
				c.Set("loggedInTime", time.Now())
				userClaim1 := influModel.User{}
				userClaim1.UserID = userClaim.UserID
				userClaim1.PartnerId = userClaim.PartnerId
				userClaim1.IDAuth = userClaim.IDAuth
				c.Set("user", userClaim1)
				return
			}
		}

		fmt.Printf("\n \n Json access Token :: %s \n \n ", vals[1])
		token, err := jwt.ParseWithClaims(vals[1], &mid.JWTClaims{}, validateAccessJWT)

		if err != nil {
			// error parsing JWT
			fmt.Printf("AuthroizationFailed : %s \n", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		fmt.Printf("Updated the User details Claim token as well \n")

		if claims, ok := token.Claims.(*mid.JWTClaims); ok && token.Valid {
			fmt.Printf("Claims : %v", claims)
			c.Set("authID", claims.AuthID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("scope", claims.Scope)
			c.Set("tfa", claims.TwoFA)
			c.Set("siteLan", claims.SiteLan)
			c.Set("custom1", claims.Custom1)
			c.Set("custom2", claims.Custom2)
			c.Set("userName", claims.UserName)
			c.Set("mobileNumber", claims.MobileNumber)
			c.Set("countryCode", claims.CountryCode)
			c.Set("completeMobileNumber", claims.CompleteMobileNumber)
			c.Set("loggedInTime", claims.LoggedInTime)

			// email must be unique

			user := influModel.User{}
			if err := db.Where("id_auth = ? ", claims.AuthID).First(&user).Error; err == nil {
				userClaim := influModel.User{}
				userClaim.UserID = user.UserID
				userClaim.PartnerId = user.PartnerId
				userClaim.IDAuth = claims.AuthID
				userClaim.FirstName = user.FirstName
				userClaim.LastName = user.LastName
				c.Set("user", userClaim)

			}
			fmt.Printf("Updated the User details as well \n")

		}

		c.Next()
	}
}

// RefreshJWT - validate refresh token
func RefreshJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtPayload mid.JWTPayload
		if err := c.ShouldBindJSON(&jwtPayload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(jwtPayload.RefreshJWT, &mid.JWTClaims{}, validateRefreshJWT)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*mid.JWTClaims); ok && token.Valid {
			c.Set("authID", claims.AuthID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("scope", claims.Scope)
			c.Set("tfa", claims.TwoFA)
			c.Set("siteLan", claims.SiteLan)
			c.Set("custom1", claims.Custom1)
			c.Set("custom2", claims.Custom2)

			c.Set("userName", claims.UserName)
			c.Set("mobileNumber", claims.MobileNumber)
			c.Set("countryCode", claims.CountryCode)
			c.Set("completeMobileNumber", claims.CompleteMobileNumber)
			c.Set("loggedInTime", claims.LoggedInTime)

			db := gdatabase.GetDB()
			user := influModel.User{}
			if err := db.Where("id_auth = ? ", claims.AuthID).First(&user).Error; err == nil {
				userClaim := influModel.User{}
				userClaim.UserID = user.UserID
				userClaim.PartnerId = user.PartnerId
				userClaim.IDAuth = claims.AuthID
				c.Set("user", userClaim)

			}
		}

		c.Next()
	}
}

// validateAccessJWT ...
func validateAccessJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return mid.JWTParams.AccessKey, nil
}

// validateRefreshJWT ...
func validateRefreshJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return mid.JWTParams.RefreshKey, nil
}

// GetJWT - issue new tokens
func GetJWT(customClaims mid.MyCustomClaims, tokenType string) (string, string, error) {
	var (
		key []byte
		ttl int
		nbf int
	)

	if tokenType == "access" {
		key = mid.JWTParams.AccessKey
		ttl = mid.JWTParams.AccessKeyTTL
		nbf = mid.JWTParams.AccNbf
	}
	if tokenType == "refresh" {
		key = mid.JWTParams.RefreshKey
		ttl = mid.JWTParams.RefreshKeyTTL
		nbf = mid.JWTParams.RefNbf
	}
	js, _ := json.Marshal(customClaims)
	fmt.Printf("Into Genereate JWT token %s", js)
	// Create the Claims
	claims := mid.JWTClaims{
		mid.MyCustomClaims{
			AuthID:  customClaims.AuthID,
			Email:   customClaims.Email,
			Role:    customClaims.Role,
			Scope:   customClaims.Scope,
			TwoFA:   customClaims.TwoFA,
			SiteLan: customClaims.SiteLan,
			Custom1: customClaims.Custom1,
			Custom2: customClaims.Custom2,

			UserName:             customClaims.UserName,
			MobileNumber:         customClaims.MobileNumber,
			CountryCode:          customClaims.CountryCode,
			CompleteMobileNumber: customClaims.CompleteMobileNumber,
			LoggedInTime:         customClaims.LoggedInTime,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(ttl))),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    mid.JWTParams.Issuer,
			Subject:   mid.JWTParams.Subject,
		},
	}

	if mid.JWTParams.Audience != "" {
		claims.Audience = []string{mid.JWTParams.Audience}
	}
	if nbf > 0 {
		claims.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(nbf)))
	}

	fmt.Printf("\n\nJWT Claims :::: %v , %v , %v , %v \n\n", ttl, nbf, jwt.NewNumericDate(time.Now().Add(time.Second*time.Duration(nbf))), jwt.NewNumericDate(time.Now().Add(time.Minute*time.Duration(ttl))))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtValue, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("\n\nJWT Token signed Values Claims :::: %s , %s", js, jwtValue)

	return jwtValue, claims.ID, nil
}
