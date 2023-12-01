package model

import (
	"errors"
	"fmt"
	"time"
)

type BotTokenModel struct {
	StringPKBaseModel
	Name         string `json:"name"`
	Token        string `json:"token"`
	ExpiresAt    int64  `json:"expiresAt"`
	RecentUsedAt int64  `json:"recentUsedAt"`
}

func (*BotTokenModel) TableName() string {
	return "bot_tokens"
}

func BotVerifyAccessToken(tokenString string) (*UserModel, error) {
	// 解析 token
	var botToken BotTokenModel
	db.Where("token = ?", tokenString).First(&botToken)

	if botToken.ID == "" {
		return nil, errors.New("bot not found")
	}

	// 验证 token 是否过期
	if botToken.ExpiresAt < time.Now().UnixMilli() {
		return nil, errors.New("token expired")
	}

	// 查询用户
	var user UserModel
	userID := botToken.ID
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("UserModel not found")
	}

	return &user, nil
}
