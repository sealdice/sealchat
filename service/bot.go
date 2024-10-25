package service

import (
	"fmt"
	"sealchat/model"
)

func BotListByChannelId(channelId string) []string {
	var ids []string
	roleId := fmt.Sprintf("ch-%s-%s", channelId, "bot")
	ids1, _ := model.UserRoleMappingUserIdListByRoleId(roleId)
	ids = append(ids, ids1...)

	ch, _ := model.ChannelGet(channelId)
	if ch.RootId != "" {
		roleId := fmt.Sprintf("ch-%s-%s", ch.RootId, "bot")
		ids2, _ := model.UserRoleMappingUserIdListByRoleId(roleId)
		ids = append(ids, ids2...)
	}

	return ids
}
