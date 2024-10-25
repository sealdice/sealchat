package pm

import "github.com/mikespook/gorbac"

var (
	PermFuncChannelRead           = gorbac.NewStdPermission("func_channel_read")             // 频道 - 查看
	PermFuncChannelTextSend       = gorbac.NewStdPermission("func_channel_text_send")        // 频道 - 文字 - 发送
	PermFuncChannelFileSend       = gorbac.NewStdPermission("func_channel_file_send")        // 频道 - 查看
	PermFuncChannelAudioSend      = gorbac.NewStdPermission("func_channel_audio_send")       // 频道 - 查看
	PermFuncChannelInvite         = gorbac.NewStdPermission("func_channel_invite")           // 频道 - 邀请
	PermFuncChannelMemberRemove   = gorbac.NewStdPermission("func_channel_member_remove")    // 频道 - 踢人
	PermFuncChannelRoleLink       = gorbac.NewStdPermission("func_channel_role_link")        // 频道 - 角色设置
	PermFuncChannelRoleUnlink     = gorbac.NewStdPermission("func_channel_role_unlink")      // 频道 - 角色设置
	PermFuncChannelRoleLinkRoot   = gorbac.NewStdPermission("func_channel_role_link_root")   // 频道 - 角色设置 (Root管理员)
	PermFuncChannelRoleUnlinkRoot = gorbac.NewStdPermission("func_channel_role_unlink_root") // 频道 - 角色设置 (Root管理员)

	PermFuncChannelReadAll     = gorbac.NewStdPermission("func_channel_read_all")      // 频道 - 查看所有子频道
	PermFuncChannelTextSendAll = gorbac.NewStdPermission("func_channel_text_send_all") // 频道 - 文字 - 发送

	PermFuncChannelSubChannelCreate = gorbac.NewStdPermission("func_channel_sub_channel_create") // 频道 - 创建子频道
)
