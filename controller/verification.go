package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/prakashjegan/golangexercise/database/model"
	"github.com/prakashjegan/golangexercise/handler"
	"github.com/prakashjegan/golangexercise/lib/renderer"
	"github.com/prakashjegan/golangexercise/service"
)

// VerifyEmail - verify email address
func VerifyEmail(c *gin.Context) {
	payload := model.AuthPayload{}
	usr, err := service.GetUserFromContext(c)
	if err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.VerifyEmail(payload, usr)

	renderer.Render(c, resp, statusCode)
}

// CreateVerificationEmail issues new verification code upon request
func CreateVerificationEmail(c *gin.Context) {
	payload := model.AuthPayload{}
	usr, err := service.GetUserFromContext(c)
	if err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateVerificationEmail(payload, usr)

	renderer.Render(c, resp, statusCode)
}

// VerifyEmail - verify email address
func VerifyMobile(c *gin.Context) {
	payload := model.AuthPayload{}
	usr, err := service.GetUserFromContext(c)
	if err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.VerifyMobile(payload, usr)

	renderer.Render(c, resp, statusCode)
}

// CreateVerificationEmail issues new verification code upon request
func CreateVerificationMobile(c *gin.Context) {
	payload := model.AuthPayload{}
	usr, err := service.GetUserFromContext(c)
	if err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	//alreadyExistingMobileNumber := c.Get("completeMobileNumber").(string)
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	//resp, statusCode := handler.CreateVerificationMobile(payload, usr)
	resp, statusCode := handler.CreateVerificationMobileViaSend(payload, usr, true)

	renderer.Render(c, resp, statusCode)
}
