// Package model contains all the models required
// for a functional database management system
package model

import (
	"time"

	"gorm.io/gorm"
)

type UserAuthTokens struct {
	UserAuthTokenID  uint64         `gorm:"primaryKey" json:"userAuthTokenID,omitempty"`
	UserID           uint64         `gorm:"userID;index" json:"userID,omitempty"`
	PartnerID        uint64         `gorm:"partnerID;index" json:"partnerID,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime;index" json:"createdAt,omitempty"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime;index" json:"updatedAt,omitempty"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Code             string         `gorm:"code;index" json:"code,omitempty"`
	AccessJWT        string         `json:"accessJWT,omitempty"`
	RefreshJWT       string         `json:"refreshJWT,omitempty"`
	UserSignUpStatus string         `json:"userSignUpStatus,omitempty"`
	Email            string         `json:"email,omitempty"`
}
