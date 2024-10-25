package model

/* 这个表格记录了用户对频道的权限情况 */

const (
	ChannelPermUserALL = "@all"
)

type ChannelPermModel struct {
	StringPKBaseModel
	ChannelID string `json:"channel_id" gorm:"index"` // 准入的频道ID
	UserID    string `json:"user_id" gorm:"index"`    // 准入的用户ID
	Role      string `json:"role"`                    // 对应权限
}

func (*ChannelPermModel) TableName() string {
	return "channel_perms"
}
