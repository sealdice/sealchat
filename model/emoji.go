package model

import (
	"gorm.io/gorm"
	"sealchat/utils"
)

type UserEmojiModel struct {
	StringPKBaseModel
	UserID       string `json:"userId" gorm:"index; index:idx_unique_user_emojis_user_id_attachment_id,unique"`
	AttachmentID string `json:"attachmentId" gorm:"index:idx_unique_user_emojis_user_id_attachment_id,unique"`
	Order        int64  `json:"order"` // 排序，默认值是时间，以后再看有啥说法
}

func (*UserEmojiModel) TableName() string {
	return "user_emojis"
}

func UserEmojiCreate(userId string, attachmentId string) (*UserEmojiModel, error) {
	item := &UserEmojiModel{
		StringPKBaseModel: StringPKBaseModel{
			ID: utils.NewID(),
		},
		UserID:       userId,
		AttachmentID: attachmentId,
	}
	return item, db.Create(item).Error
}

func UserEmojiList(userId string, page, pageSize int) ([]*UserEmojiModel, int64, error) {
	return utils.QueryPaginatedList(db, page, pageSize, &UserEmojiModel{}, func(q *gorm.DB) *gorm.DB {
		return q.Where("user_id = ?", userId)
	})
}
