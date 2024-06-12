package model

import (
	"encoding/hex"
	"encoding/json"
	"sealchat/utils"

	"gorm.io/gorm"
)

type ByteArray []byte

func (m ByteArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(m))
}

type Attachment struct {
	StringPKBaseModel
	Hash      ByteArray `gorm:"index" json:"hash"`
	Filename  string    `json:"filename"`
	Size      int64     `gorm:"index" json:"size"`
	UserID    string    `json:"userId" gorm:"index"`
	ChannelID string    `json:"channel_id"` // 上传的频道ID
}

func (*Attachment) TableName() string {
	return "attachments"
}

func AttachmentCreate(at *Attachment) (tx *gorm.DB, item *Attachment) {
	db := GetDB()
	at.ID = utils.NewID()
	return db.Create(at), at
}
