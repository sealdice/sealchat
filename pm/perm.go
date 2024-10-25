package pm

import (
	"errors"
	"fmt"
	"github.com/mikespook/gorbac"
	"github.com/samber/lo"
	"log"
	"sealchat/model"
	"sealchat/utils"
)

const (
	// PermUnset 未设定
	PermUnset = 0
	// PermAllowed 许可
	PermAllowed = 1
	// PermDenied 禁止
	PermDenied = 2
)

var perm *gorbac.RBAC

func GetPerm() *gorbac.RBAC {
	return perm
}

// func Can(role string, permission gorbac.Permission) bool {
// 	return perm.IsGranted(role, permission, nil)
// }

func canBase(uid, channelId, roleType string, permissions ...gorbac.Permission) bool {
	roles, _ := model.UserRoleMappingListByUserID(uid, channelId, roleType)

	if channelId != "" {
		ch, _ := model.ChannelGet(channelId)
		if ch.PermType == "public" {
			roleId := fmt.Sprintf("ch-%s-%s", channelId, "visitor")
			roles = append(roles, roleId)
		}
	}

	for _, permission := range permissions {
		if gorbac.AnyGranted(perm, roles, permission, nil) {
			return true
		}
	}
	return false
}

func Can(uid string, channelId string, permissions ...gorbac.Permission) bool {
	return canBase(uid, channelId, "", permissions...)
}

func CanWithSystemRole(uid string, permissions ...gorbac.Permission) bool {
	return canBase(uid, "", "system", permissions...)
}

func CanWithChannelRole(uid string, channelId string, permissions ...gorbac.Permission) bool {
	ch, _ := model.ChannelGet(channelId)
	if ch.RootId != "" {
		// 如果是子频道，若根节点频道允许，则同时允许
		var newPerm []gorbac.Permission
		for _, i := range permissions {
			switch i.ID() {
			case PermFuncChannelReadAll.ID(), PermFuncChannelTextSendAll.ID():
				newPerm = append(newPerm, i)
			}
		}
		if ok := canBase(uid, ch.RootId, "channel", newPerm...); ok {
			return true
		}
	}

	return canBase(uid, channelId, "channel", permissions...)
}

func sysRolesInit() {
	roleAdmin := gorbac.NewStdRole("sys-admin")
	_ = roleAdmin.Assign(PermModAdmin)
	_ = roleAdmin.Assign(PermFuncAdminServeConfig)
	_ = roleAdmin.Assign(PermFuncAdminBotTokenView)
	_ = roleAdmin.Assign(PermFuncAdminBotTokenCreate)
	_ = roleAdmin.Assign(PermFuncAdminBotTokenEdit)
	_ = roleAdmin.Assign(PermFuncAdminBotTokenDelete)
	_ = roleAdmin.Assign(PermFuncAdminUserSetEnable)
	_ = roleAdmin.Assign(PermFuncAdminUserPasswordReset)
	_ = roleAdmin.Assign(PermFuncAdminUserEdit)
	_ = roleAdmin.Assign(PermFuncChannelCreatePublic)
	_ = roleAdmin.Assign(PermFuncChannelCreateNonPublic)

	roleUser := gorbac.NewStdRole("sys-user")
	_ = roleUser.Assign(PermFuncChannelCreateNonPublic)

	roleVisitor := gorbac.NewStdRole("sys-visitor")

	_ = perm.Add(roleAdmin)
	_ = perm.Add(roleUser)
	_ = perm.Add(roleVisitor)
}

func Init() {
	perm = gorbac.New()
	sysRoles, num, _ := model.SystemRoleList(0, -1)
	chRoles, _, _ := model.ChannelRoleAllList(0, -1)

	if num == 0 {
		// 目前system roles表还未实用，每次创建是设计的一部分而不是bug
		sysRolesInit()
	} else {
		for _, i := range sysRoles {
			lst, _ := model.RolePermissionList(i.ID)
			role := gorbac.NewStdRole(i.ID)

			for _, j := range lst {
				_ = role.Assign(gorbac.NewStdPermission(j))
			}
			_ = perm.Add(role)
		}
	}

	for _, i := range chRoles {
		lst, _ := model.RolePermissionList(i.ID)
		role := gorbac.NewStdRole(i.ID)

		for _, j := range lst {
			_ = role.Assign(gorbac.NewStdPermission(j))
		}
		_ = perm.Add(role)
	}
}

func GetAllSysPermByUid(uid string) *utils.SyncSet[string] {
	roles, _ := model.UserRoleMappingListByUserID(uid, "", "system")
	permSet := &utils.SyncSet[string]{}

	for _, name := range roles {
		r, _, err := perm.Get(name)
		if err != nil {
			continue
		}
		if x, ok := r.(*gorbac.StdRole); ok {
			lo.Map(x.Permissions(), func(item gorbac.Permission, index int) string {
				key := item.ID()
				permSet.Add(key)
				return key
			})
		}
	}

	return permSet
}

func ChannelRoleSet(roleId string, perms []gorbac.Permission) {
	roleCur := gorbac.NewStdRole(roleId)
	for _, perm := range perms {
		if err := roleCur.Assign(perm); err != nil {
			log.Printf("分配权限失败: %v", err)
		}
	}

	if _, _, err := perm.Get(roleId); !errors.Is(err, gorbac.ErrRoleNotExist) {
		_ = perm.Remove(roleId)
	}

	if err := perm.Add(roleCur); err != nil {
		log.Printf("添加角色到RBAC系统失败: %v", err)
	}
}
