package service

import (
	"internhub/user-service/internal/model"
	"internhub/user-service/internal/repository"
)

func GetProfile(userID uint) (*model.UserProfile, error) {
	return repository.GetProfile(userID)
}

func UpdateProfile(userID uint, nickname, avatar string) (*model.UserProfile, error) {
	p, _ := repository.GetProfile(userID)
	if p == nil {
		p = &model.UserProfile{UserID: userID}
	}
	if nickname != "" {
		p.Nickname = nickname
	}
	if avatar != "" {
		p.Avatar = avatar
	}
	if err := repository.UpsertProfile(p); err != nil {
		return nil, err
	}
	return p, nil
}
