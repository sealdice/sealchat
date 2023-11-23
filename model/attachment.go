package model

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

type Attachment struct {
	StringPKBaseModel
	Hash      []byte `gorm:"index" json:"hash"`
	Filename  string `json:"filename"`
	Size      int64  `gorm:"index" json:"size"`
	UserID    string `json:"userId" gorm:"index"`
	ChannelID string `json:"channel_id"` // 上传的频道ID
}

func (*Attachment) TableName() string {
	return "attachments"
}

func AttachmentCreate(at *Attachment) (tx *gorm.DB) {
	db := GetDB()
	at.ID = gonanoid.Must()
	return db.Create(at)
}
