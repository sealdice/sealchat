package model

import (
	"time"

	"sealchat/protocol"
)

type MessageModel struct {
	StringPKBaseModel
	Content   string `json:"content"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id" gorm:"null;size:100"`
	MemberID  string `json:"member_id" gorm:"null;size:100"`
	UserID    string `json:"user_id" gorm:"null;size:100"`
	QuoteID   string `json:"quote_id" gorm:"null;size:100"`

	IsRevoked bool `json:"is_revoked" gorm:"null"` // 被撤回。这样实现可能不很严肃，但是能填补窗口中空白

	SenderMemberName string `json:"sender_member_name"` // 用户在当时的名字

	User   *UserModel    `json:"user"`   // 嵌套 User 结构体
	Member *MemberModel  `json:"member"` // 嵌套 Member 结构体
	Quote  *MessageModel `json:"quote"`  // 嵌套 Message 结构体
}

func (*MessageModel) TableName() string {
	return "messages"
}

func (m *MessageModel) ToProtocolType2(channelData *protocol.Channel) *protocol.Message {
	return &protocol.Message{
		ID:      m.ID,
		Content: m.Content,
		Channel: channelData,
		// User:      userData,
		// Member:    member.ToProtocolType(),
		CreatedAt: time.Now().UnixMilli(), // 跟js相匹配
	}
}
