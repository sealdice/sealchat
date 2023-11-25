package model

type MentionModel struct {
	StringPKBaseModel
	ReceiverId string `gorm:"index" json:"receiverId"` // 被提及人
	SenderId   string `gorm:"index" json:"senderId"`   // 发起人

	LocPostType string `gorm:"index" json:"locPostType"` // 事发地对象类型
	LocPostID   string `gorm:"index" json:"locPostId"`   // 事发地对象ID

	RelatedType string `gorm:"index" json:"relatedType"` // 消息类型
	RelatedID   string `gorm:"index" json:"relatedId"`   // 消息ID

	Data ByteArray `gorm:"null" json:"data"` // 附加数据，一般用不着
}

func (*MentionModel) TableName() string {
	return "mentions"
}
