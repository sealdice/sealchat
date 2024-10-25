package model

import (
	"time"

	"sealchat/protocol"
)

type MemberModel struct {
	StringPKBaseModel
	Nickname     string `gorm:"null" json:"nick"`           // 昵称
	ChannelID    string `gorm:"not null" json:"channel_id"` // 频道ID
	UserID       string `json:"user_id" gorm:"index,null"`  // 用户ID
	RecentSentAt int64  `json:"recentSentAt"`               // 最近发送消息的时间
}

func (u *MemberModel) SaveInfo() {
	db.Model(u).Select("nickname").Updates(u)
}

func (*MemberModel) TableName() string {
	return "members"
}

func (u *MemberModel) ToProtocolType() *protocol.GuildMember {
	return &protocol.GuildMember{
		ID: u.ID,
		User: &protocol.User{
			ID: u.UserID,
		},
		Nick: u.Nickname,
	}
}

func (m *MemberModel) UpdateRecentSent() {
	m.RecentSentAt = time.Now().UnixMilli()
	db.Model(m).Update("recent_sent_at", m.RecentSentAt)
}

func MemberGetByUserIDAndChannelIDBase(userId string, channelId string, defaultName string, createIfNotExists bool) (*MemberModel, error) {
	var member MemberModel
	err := db.Where("user_id = ? AND channel_id = ?", userId, channelId).Limit(1).Find(&member).Error
	if member.ID == "" {
		// 未找到记录，尝试创建新的记录
		if createIfNotExists {
			x := MemberModel{UserID: userId, ChannelID: channelId, Nickname: defaultName}
			err = db.Create(&x).Error
			return &x, err
		}
		return nil, nil
	}
	return &member, nil
}

func MemberGetByUserIDAndChannelID(userId string, channelId string, defaultName string) (*MemberModel, error) {
	return MemberGetByUserIDAndChannelIDBase(userId, channelId, defaultName, true)
}
