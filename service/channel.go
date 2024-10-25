package service

import (
	"fmt"
	"github.com/mikespook/gorbac"
	"strings"

	"github.com/samber/lo"

	"sealchat/model"
	"sealchat/pm"
)

// ChannelList 获取可见的频道
func ChannelList(userId string) ([]*model.ChannelModel, error) {
	// 包括如下内容:
	// 1. 属性为可见的一级频道(即没有父级的频道)
	// 2. 具有明确可看权限的频道(先查频道角色，再根据频道角色验证权限和获取频道id)
	// 3. 补入有权查看的频道的子频道

	roles, err := model.UserRoleMappingListByUserID(userId, "", "channel")
	if err != nil {
		return nil, err
	}

	var rolesCanRead []string
	db := model.GetDB()
	db.Model(&model.RolePermissionModel{}).
		Where("role_id in ? and permission_id in ?", roles, []string{pm.PermFuncChannelRead.ID(), pm.PermFuncChannelReadAll.ID()}).
		Pluck("role_id", &rolesCanRead)

	// 这里获得的是2
	ids2 := lo.Map(rolesCanRead, func(item string, index int) string {
		return strings.SplitN(item, "-", 3)[1]
	})

	// 3.1，公开子频道
	var ids3 []string
	db.Model(&model.ChannelModel{}).Where("root_id in ? and perm_type = ?", ids2, "public").
		Pluck("id", &ids3)

	// 3.2
	// 先找出我有“查看全部”权限的的顶级频道
	// 找出这些顶级频道的下属频道
	var rolesCanRead2 []string
	db.Model(&model.RolePermissionModel{}).
		Where("role_id in ? and permission_id in ?", roles, []string{pm.PermFuncChannelReadAll.ID()}).
		Pluck("role_id", &rolesCanRead2)
	ids2x := lo.Map(rolesCanRead2, func(item string, index int) string {
		return strings.SplitN(item, "-", 3)[1]
	})
	var ids32 []string
	db.Model(&model.ChannelModel{}).Where("root_id in ? and perm_type = ?", ids2x, "non-public").
		Pluck("id", &ids32)

	idsAll := append(ids2, ids3...)
	idsAll = append(idsAll, ids32...)
	var items []*model.ChannelModel
	db.Model(&model.ChannelModel{}).Where("id in ? or perm_type = ?", idsAll, "public").
		Group("id").              // 使用Group来去重
		Order("created_at DESC"). // 按创建时间降序排列
		Find(&items)

	return items, nil
}

func ChannelNew(channelID, channelType, channelName string, creatorId string, parentId string) *model.ChannelModel {
	m := model.ChannelPublicNew(channelID, &model.ChannelModel{
		Name:     channelName,
		PermType: channelType,
		ParentID: parentId,
		RootId:   parentId, // TODO: 这个是不准的，但是目前不允许二级以上子频道
	}, creatorId)

	roleCreate(channelID, "owner", "群主", func(roleId string) []gorbac.Permission {
		return []gorbac.Permission{
			pm.PermFuncChannelRead,
			pm.PermFuncChannelTextSend,
			pm.PermFuncChannelFileSend,
			pm.PermFuncChannelAudioSend,
			pm.PermFuncChannelInvite,
			// pm.PermFuncChannelMemberRemove,
			pm.PermFuncChannelSubChannelCreate,
			pm.PermFuncChannelRoleLink,
			pm.PermFuncChannelRoleUnlink,
			pm.PermFuncChannelRoleLinkRoot,
			pm.PermFuncChannelRoleUnlinkRoot,
			pm.PermFuncChannelReadAll,
			pm.PermFuncChannelTextSendAll,
		}
	})

	roleCreate(channelID, "admin", "管理员", func(roleId string) []gorbac.Permission {
		return []gorbac.Permission{
			pm.PermFuncChannelRead,
			pm.PermFuncChannelTextSend,
			pm.PermFuncChannelFileSend,
			pm.PermFuncChannelAudioSend,
			pm.PermFuncChannelInvite,
			// pm.PermFuncChannelMemberRemove,
			pm.PermFuncChannelSubChannelCreate,
			pm.PermFuncChannelRoleLink,
			pm.PermFuncChannelRoleUnlink,
			pm.PermFuncChannelReadAll,
			pm.PermFuncChannelTextSendAll,
		}
	})

	roleCreate(channelID, "ob", "观察者", func(roleId string) []gorbac.Permission {
		return []gorbac.Permission{
			pm.PermFuncChannelRead,
			pm.PermFuncChannelTextSend,
			pm.PermFuncChannelFileSend,
			pm.PermFuncChannelAudioSend,
			pm.PermFuncChannelReadAll,
		}
	})

	roleCreate(channelID, "visitor", "游客", func(roleId string) []gorbac.Permission {
		return []gorbac.Permission{
			pm.PermFuncChannelRead,
			pm.PermFuncChannelTextSend,
		}
	})

	roleCreate(channelID, "member", "成员", func(roleId string) []gorbac.Permission {
		return []gorbac.Permission{
			pm.PermFuncChannelRead,
			pm.PermFuncChannelTextSend,
			pm.PermFuncChannelFileSend,
			pm.PermFuncChannelAudioSend,
			pm.PermFuncChannelInvite,
		}
	})

	roleCreate(channelID, "bot", "机器人", func(roleId string) []gorbac.Permission {
		return []gorbac.Permission{
			pm.PermFuncChannelReadAll,
			pm.PermFuncChannelTextSendAll,
		}
	})

	roleId := fmt.Sprintf("ch-%s-%s", channelID, "owner")
	_ = model.UserRoleMappingCreate(&model.UserRoleMappingModel{
		UserID:   creatorId,
		RoleID:   roleId,
		RoleType: "channel",
	})

	return m
}
