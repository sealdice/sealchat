package api

import (
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
)

func (ctx *ChatContext) TagCheck(ChannelID, msgId, text string) {
	root := protocol.ElementParse(text)
	db := model.GetDB()

	root.Traverse(func(el *protocol.Element) {
		switch el.Type {
		case "at":
			mention := model.MentionModel{
				StringPKBaseModel: model.StringPKBaseModel{
					ID: utils.NewID(),
				},
				SenderId:    ctx.User.ID,
				LocPostType: "channel",
				LocPostID:   ChannelID,
				RelatedType: "message",
				RelatedID:   msgId,
			}
			if el.Attrs["role"] == "all" {
				mention.ReceiverId = "all"
				db.Create(&mention)
			} else {
				if id, exists := el.Attrs["id"]; exists {
					mention.ReceiverId = id.(string)
					db.Create(&mention)
				}
			}
		}
	})
	// fmt.Println("xxx", root.ToString())
}
