package model

type ChannelModel struct {
	StringPKBaseModel
	Name     string `json:"name"`
	ParentID string `json:"parent_id" gorm:"null"`
}

func (*ChannelModel) TableName() string {
	return "channels"
}
