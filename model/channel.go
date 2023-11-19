package model

type ChannelModel struct {
	StringPKBaseModel
	Name         string `json:"name"`
	ParentID     string `json:"parentId" gorm:"null"` // 好像satori协议这里不统一啊
	MembersCount int    `json:"membersCount" gorm:"-"`
}

func (*ChannelModel) TableName() string {
	return "channels"
}
