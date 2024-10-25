package pm

import "github.com/mikespook/gorbac"

// 全局权限

// 权限相关名词表: view create edit del export import upload

var (
	PermModAdmin             = gorbac.NewStdPermission("mod_admin")               // 查看管理界面
	PermFuncAdminServeConfig = gorbac.NewStdPermission("func_admin_serve_config") // 修改serve配置

	PermFuncAdminBotTokenView   = gorbac.NewStdPermission("func_admin_bot_token_view")   // 查看机器人令牌
	PermFuncAdminBotTokenCreate = gorbac.NewStdPermission("func_admin_bot_token_create") // 创建机器人令牌
	PermFuncAdminBotTokenEdit   = gorbac.NewStdPermission("func_admin_bot_token_edit")   // 编辑机器人令牌
	PermFuncAdminBotTokenDelete = gorbac.NewStdPermission("func_admin_bot_token_delete") // 删除机器人令牌

	PermFuncAdminUserSetEnable     = gorbac.NewStdPermission("func_admin_user_set_enable")     // 禁用和启用用户
	PermFuncAdminUserPasswordReset = gorbac.NewStdPermission("func_admin_user_password_reset") // 重置用户密码
	PermFuncAdminUserEdit          = gorbac.NewStdPermission("func_admin_user_edit")           // 设置角色等

	PermFuncChannelCreatePublic    = gorbac.NewStdPermission("func_channel_create_public")     // 频道 - 创建公开频道
	PermFuncChannelCreateNonPublic = gorbac.NewStdPermission("func_channel_create_non_public") // 频道 - 创建个人频道

	// "g_channel_settings_edit": 1, // 频道管理 - 通用权限
)
