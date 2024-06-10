package model

import (
	"sealchat/utils"
)

type UserEmojiModel struct {
	StringPKBaseModel
	UserID       string `json:"userId" gorm:"index; uniqueIndex:idx_unique_user_emojis_user_id_attachment_id"`
	AttachmentID string `json:"attachmentId" gorm:"uniqueIndex:idx_unique_user_emojis_user_id_attachment_id"`
	Order        int64  `json:"order"` // 排序，默认值是时间，以后再看有啥说法
}

func (*UserEmojiModel) TableName() string {
	return "user_emojis"
}

func UserEmojiCreate(userId string, attachmentId string) error {
	return db.Create(&UserEmojiModel{
		StringPKBaseModel: StringPKBaseModel{
			ID: utils.NewID(),
		},
		UserID:       userId,
		AttachmentID: attachmentId,
	}).Error
}
