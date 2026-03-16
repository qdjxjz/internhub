package repository

import (
	"internhub/user-service/config"
	"internhub/user-service/internal/model"
)

func GetProfile(userID uint) (*model.UserProfile, error) {
	var p model.UserProfile
	err := config.DB.Where("user_id = ?", userID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func UpsertProfile(p *model.UserProfile) error {
	return config.DB.Save(p).Error
}
