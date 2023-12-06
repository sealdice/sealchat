package model

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/blake2s"
	"sealchat/protocol"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// UserModel 用户表
type UserModel struct {
	StringPKBaseModel
	Nickname string `gorm:"null" json:"nick"` // 昵称
	Avatar   string `json:"avatar"`           // 头像
	Brief    string `json:"brief"`            // 简介
	Role     string `json:"role"`             // 权限

	Username string `gorm:"uniqueIndex;not null" json:"username"` // 用户名，唯一，非空
	Password string `gorm:"not null" json:"-"`                    // 密码，非空
	Salt     string `gorm:"not null" json:"-"`                    // 盐，非空
	IsBot    bool   `gorm:"null" json:"is_bot"`                   // 是否是机器人

	// Token          string `gorm:"index" json:"token"` // 令牌
	// TokenExpiresAt int64  `json:"expiresAt"`
	// RecentSentAt int64 `json:"recentSentAt"` // 最近发送消息的时间
}

func (*UserModel) TableName() string {
	return "users"
}

func (u *UserModel) ToProtocolType() *protocol.User {
	return &protocol.User{
		ID:     u.ID,
		Nick:   u.Nickname,
		Avatar: u.Avatar,
		IsBot:  u.IsBot,
	}
}

func (u *UserModel) SaveAvatar() {
	db.Model(u).Update("avatar", u.Avatar)
}

func (u *UserModel) SaveInfo() {
	db.Model(u).Select("nickname", "brief").Updates(u)
}

// AccessTokenModel access_token表
type AccessTokenModel struct {
	StringPKBaseModel
	UserID    string    `gorm:"not null"`             // 用户ID，非空
	Token     string    `gorm:"uniqueIndex;not null"` // token，唯一，非空
	ExpiredAt time.Time `gorm:"not null"`             // 过期时间，非空
}

func (*AccessTokenModel) TableName() string {
	return "access_tokens"
}

// 生成随机盐
func generateSalt() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return ""
	}
	return base64.RawStdEncoding.EncodeToString(salt)
}

// 使用盐对密码进行哈希
func hashPassword(password string, salt string) (string, error) {
	// 将密码和盐拼接起来
	saltedPassword := password + salt

	// 计算哈希值
	hashBytes := blake2s.Sum256([]byte(saltedPassword))
	hash := base64.RawStdEncoding.EncodeToString(hashBytes[:])

	return hash, nil
}

// 创建用户
func UserCreate(username, password string, nickname string) (*UserModel, error) {
	var role string

	var count int64
	db.Select("id").Find(&UserModel{}).Count(&count)
	if count == 0 {
		role = "role-admin"
	}

	salt := generateSalt()
	hashedPassword, err := hashPassword(password, salt)
	if err != nil {
		return nil, err
	}
	user := &UserModel{
		Role:     role,
		Username: username,
		Nickname: nickname,
		Password: hashedPassword,
		Salt:     salt,
	}
	user.ID = gonanoid.Must()
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// 登录认证
func UserAuthenticate(username, password string) (*UserModel, error) {
	var user UserModel
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	hashedPassword, err := hashPassword(password, user.Salt)
	if err != nil {
		return nil, err
	}
	if hashedPassword != user.Password {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}

var JWT_KEY = []byte("sc-jwt_secret")

func AcessTokenDeleteAllByUserID(userID string) error {
	return db.Where("user_id = ?", userID).Delete(&AccessTokenModel{}).Error
}

// 生成 access_token
func UserGenerateAccessToken(userID string) (string, error) {
	expiredAt := time.Now().Add(time.Duration(15*24) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    expiredAt.Unix(),
	})
	signedToken, err := token.SignedString(JWT_KEY)
	if err != nil {
		return "", err
	}
	accessToken := &AccessTokenModel{
		UserID:    userID,
		Token:     signedToken,
		ExpiredAt: expiredAt,
	}
	accessToken.ID = gonanoid.Must()
	if err := db.Create(accessToken).Error; err != nil {
		return "", err
	}
	return signedToken, nil
}

// 验证 access_token 是否有效
func UserVerifyAccessToken(tokenString string) (*UserModel, error) {
	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// 返回用于签名的密钥
		return []byte(JWT_KEY), nil
	})
	if err != nil {
		return nil, err
	}

	// 获取 token 中的用户 ID
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		return nil, fmt.Errorf("Invalid user ID")
	}

	// 查询用户
	var user UserModel
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("UserModel not found")
	}

	// 验证 token 是否过期
	expirationTime, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("Invalid token expiration time")
	}
	if time.Now().Unix() > int64(expirationTime) {
		return nil, fmt.Errorf("Token expired")
	}

	return &user, nil
}

// 修改密码
func UserUpdatePassword(userID string, newPassword string) error {
	// 查询用户
	var user UserModel
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("UserModel not found")
	}

	// 更新密码
	salt := generateSalt()
	hashedNewPassword, err := hashPassword(newPassword, salt)
	if err != nil {
		return err
	}
	if err := db.Model(&user).Updates(map[string]interface{}{
		"password": hashedNewPassword,
		"salt":     salt,
	}).Error; err != nil {
		return err
	}
	return nil
}
