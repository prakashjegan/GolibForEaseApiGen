// Package service contains common functions used by
// the whole application
package service

import (
	"github.com/prakashjegan/golangexercise/database"
	"github.com/prakashjegan/golangexercise/database/model"
	influmodel "github.com/prakashjegan/golangexercise/golangexercise-service/database/model"
	"github.com/prakashjegan/golangexercise/golangexercise-service/utils"
)

// GetUserByEmail ...
func GetUserByEmail(email string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("email = ? ", utils.DecryptDataWithOutError(email)).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}

func GetUserByUserName(userName string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("user_name = ? ", userName).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}

func GetUserByCompleteMobileNumber(completeMobileNumber string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("complete_mobile_number = ? ", utils.DecryptDataWithOutError(completeMobileNumber)).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}

func GetUserById(userId uint64) (*influmodel.User, error) {
	db := database.GetDB()

	var auth influmodel.User

	if err := db.Where("user_id = ? ", userId).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}
