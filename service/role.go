package service

import (
	"fmt"
	"log"

	"github.com/mikespook/gorbac"

	"sealchat/model"
	"sealchat/pm"
)

func roleCreate(channelID, key string, name string, f func(role string) []gorbac.Permission) *model.ChannelRoleModel {
	cr := &model.ChannelRoleModel{}
	roleId := fmt.Sprintf("ch-%s-%s", channelID, key)
	cr.ID = roleId

	perms := f(roleId)

	permIDs := make([]string, len(perms))
	for i, perm := range perms {
		permIDs[i] = perm.ID()
	}

	err := model.ChannelRoleCreate(&model.ChannelRoleModel{
		StringPKBaseModel: model.StringPKBaseModel{ID: roleId},
		Name:              name,
		ChannelID:         channelID,
	})

	if err != nil {
		log.Printf("创建角色权限失败[步骤1]: %v", err)
		return nil
	}

	if err := model.RolePermissionBatchCreate(roleId, permIDs); err != nil {
		log.Printf("创建角色权限失败[步骤2]: %v", err)
		return nil
	}

	pm.ChannelRoleSet(roleId, perms)

	return cr
}

func UserRoleUnlink(roleIds []string, userIds []string) (int64, error) {
	num, err := model.UserRoleUnlink(roleIds, userIds)
	// TODO: 做一些特殊处理，比如说阻止用户自我删除之类
	return num, err
}

func UserRoleLink(roleIds []string, userIds []string) (int64, error) {
	num, err := model.UserRoleLink(roleIds, userIds)
	// 也是特殊处理
	return num, err
}
