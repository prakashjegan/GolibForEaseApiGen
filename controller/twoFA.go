package controller

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/prakashjegan/golangexercise/database/model"
	"github.com/prakashjegan/golangexercise/handler"
	"github.com/prakashjegan/golangexercise/lib/renderer"
	"github.com/prakashjegan/golangexercise/service"
)

// Setup2FA - get secret to activate 2FA
// possible for accounts without 2FA-ON
func Setup2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Setup2FA(claims, password)

	if statusCode != 201 {
		renderer.Render(c, resp, statusCode)
		return
	}

	// serve the QR code
	c.File(fmt.Sprintf("%v", resp.Message))
}

// Activate2FA - activate 2FA upon validation
// possible for accounts without 2FA-ON
func Activate2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	otp := model.AuthPayload{}
	if err := c.ShouldBindJSON(&otp); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Activate2FA(claims, otp)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Validate2FA - issue new JWTs upon 2FA validation
// required for accounts with 2FA-ON
func Validate2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	otp := model.AuthPayload{}
	if err := c.ShouldBindJSON(&otp); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Validate2FA(claims, otp)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Deactivate2FA - disable 2FA for user account
func Deactivate2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Deactivate2FA(claims, password)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}
