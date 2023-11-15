package model

type GuildModel struct {
	StringPKBaseModel
	Name   string `json:"name"`   // 群组名称
	Avatar string `json:"avatar"` // 群组头像
}

func (*GuildModel) TableName() string {
	return "guilds"
}
