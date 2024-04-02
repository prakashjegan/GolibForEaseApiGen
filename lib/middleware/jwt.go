package middleware

// github.com/prakashjegan/golangexercise
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	// "fmt"
	// "net/http"
	// "strings"
	// "time"

	// "github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	// "github.com/google/uuid"
	// gdatabase "github.com/prakashjegan/golangexercise/database"
	// influModel "github.com/prakashjegan/golangexercise/golangexercise-service/database/model"
)

// JWTParameters - params to configure JWT
type JWTParameters struct {
	AccessKey     []byte
	AccessKeyTTL  int
	RefreshKey    []byte
	RefreshKeyTTL int

	Audience string
	Issuer   string
	AccNbf   int
	RefNbf   int
	Subject  string
}

// JWTParams - exported variables
var JWTParams JWTParameters

// MyCustomClaims ...
type MyCustomClaims struct {
	AuthID               uint64 `json:"authID,omitempty"`
	Email                string `json:"email,omitempty"`
	UserName             string `json:"userName,omitempty"`
	MobileNumber         string `gorm:"index" json:"mobileNumber,omitempty"`
	CountryCode          string `json:"countryCode,omitempty"`
	CompleteMobileNumber string `gorm:"index" json:"completeMobileNumber,omitempty"`
	UserID               uint64 `gorm:"index" json:"userId,omitempty"`
	Role                 string `json:"role,omitempty"`
	Scope                string `json:"scope,omitempty"`
	TwoFA                string `json:"twoFA,omitempty"`
	SiteLan              string `json:"siteLan,omitempty"`
	Custom1              string `json:"custom1,omitempty"`
	Custom2              string `json:"custom2,omitempty"`
	LoggedInTime         uint64 `json:"loggedInTime,omitempty"`
}

// JWTClaims ...
type JWTClaims struct {
	MyCustomClaims
	jwt.RegisteredClaims
}

// JWTPayload ...
type JWTPayload struct {
	AccessJWT        string `json:"accessJWT,omitempty"`
	RefreshJWT       string `json:"refreshJWT,omitempty"`
	Email            string `json:"email,omitempty"`
	UserName         string `json:"userName,omitempty"`
	MobileNumber     string `json:"mobileNumber,omitempty"`
	UserSignUpStatus string `json:"userSignUpStatus,omitempty"`
}

// // JWT - validate access token
// func JWT() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		val := c.Request.Header.Get("Authorization")
// 		if len(val) == 0 || !strings.Contains(val, "Bearer ") {
// 			// no vals or no bearer found
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}
// 		vals := strings.Split(val, " ")
// 		if len(vals) != 2 {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		token, err := jwt.ParseWithClaims(vals[1], &JWTClaims{}, validateAccessJWT)

// 		if err != nil {
// 			// error parsing JWT
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
// 			c.Set("authID", claims.AuthID)
// 			c.Set("email", claims.Email)
// 			c.Set("role", claims.Role)
// 			c.Set("scope", claims.Scope)
// 			c.Set("tfa", claims.TwoFA)
// 			c.Set("siteLan", claims.SiteLan)
// 			c.Set("custom1", claims.Custom1)
// 			c.Set("custom2", claims.Custom2)
// 			c.Set("userName", claims.UserName)
// 			c.Set("mobileNumber", claims.MobileNumber)
// 			c.Set("countryCode", claims.CountryCode)
// 			c.Set("completeMobileNumber", claims.CompleteMobileNumber)
// 			c.Set("loggedInTime", claims.LoggedInTime)

// 			// email must be unique
// 			db := gdatabase.GetDB()
// 			user := influModel.Users{}
// 			if err := db.Where("id_auth = ? ", claims.AuthID).First(&user).Error; err == nil {
// 				httpResponse.Message = "mobileNumber already registered"
// 				httpStatusCode = http.StatusForbidden
// 				return
// 			}

// 		}

// 		c.Next()
// 	}
// }

// // RefreshJWT - validate refresh token
// func RefreshJWT() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var jwtPayload JWTPayload
// 		if err := c.ShouldBindJSON( c , &jwtPayload); err != nil {
// 			c.AbortWithStatus(http.StatusBadRequest)
// 			return
// 		}

// 		token, err := jwt.ParseWithClaims(jwtPayload.RefreshJWT, &JWTClaims{}, validateRefreshJWT)
// 		if err != nil {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
// 			c.Set("authID", claims.AuthID)
// 			c.Set("email", claims.Email)
// 			c.Set("role", claims.Role)
// 			c.Set("scope", claims.Scope)
// 			c.Set("tfa", claims.TwoFA)
// 			c.Set("siteLan", claims.SiteLan)
// 			c.Set("custom1", claims.Custom1)
// 			c.Set("custom2", claims.Custom2)

// 			c.Set("userName", claims.UserName)
// 			c.Set("mobileNumber", claims.MobileNumber)
// 			c.Set("countryCode", claims.CountryCode)
// 			c.Set("completeMobileNumber", claims.CompleteMobileNumber)
// 			c.Set("loggedInTime", claims.LoggedInTime)
// 		}

// 		c.Next()
// 	}
// }

// // validateAccessJWT ...
// func validateAccessJWT(token *jwt.Token) (interface{}, error) {
// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 	}
// 	return JWTParams.AccessKey, nil
// }

// // validateRefreshJWT ...
// func validateRefreshJWT(token *jwt.Token) (interface{}, error) {
// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 	}
// 	return JWTParams.RefreshKey, nil
// }

// // GetJWT - issue new tokens
// func GetJWT(customClaims MyCustomClaims, tokenType string) (string, string, error) {
// 	var (
// 		key []byte
// 		ttl int
// 		nbf int
// 	)

// 	if tokenType == "access" {
// 		key = JWTParams.AccessKey
// 		ttl = JWTParams.AccessKeyTTL
// 		nbf = JWTParams.AccNbf
// 	}
// 	if tokenType == "refresh" {
// 		key = JWTParams.RefreshKey
// 		ttl = JWTParams.RefreshKeyTTL
// 		nbf = JWTParams.RefNbf
// 	}
// 	// Create the Claims
// 	claims := JWTClaims{
// 		MyCustomClaims{
// 			AuthID:  customClaims.AuthID,
// 			Email:   customClaims.Email,
// 			Role:    customClaims.Role,
// 			Scope:   customClaims.Scope,
// 			TwoFA:   customClaims.TwoFA,
// 			SiteLan: customClaims.SiteLan,
// 			Custom1: customClaims.Custom1,
// 			Custom2: customClaims.Custom2,

// 			UserName:             customClaims.UserName,
// 			MobileNumber:         customClaims.MobileNumber,
// 			CountryCode:          customClaims.CountryCode,
// 			CompleteMobileNumber: customClaims.CompleteMobileNumber,
// 			LoggedInTime:         customClaims.LoggedInTime,
// 		},
// 		jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(ttl))),
// 			ID:        uuid.NewString(),
// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 			Issuer:    JWTParams.Issuer,
// 			Subject:   JWTParams.Subject,
// 		},
// 	}

// 	if JWTParams.Audience != "" {
// 		claims.Audience = []string{JWTParams.Audience}
// 	}
// 	if nbf > 0 {
// 		claims.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(nbf)))
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	jwtValue, err := token.SignedString(key)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return jwtValue, claims.ID, nil
// }
