
export enum PermResult {
  UNSET = 0,
  ALLOWED = 1,
  DENIED = 2
}

export interface SystemRolePermSheet {
  mod_admin: PermResult; // 查看管理界面

  func_admin_serve_config: PermResult; // 修改serve配置
  func_admin_bot_token_view: PermResult; // 查看机器人令牌
  func_admin_bot_token_create: PermResult; // 创建机器人令牌
  func_admin_bot_token_edit: PermResult; // 编辑机器人令牌
  func_admin_bot_token_delete: PermResult; // 删除机器人令牌

  func_admin_user_set_enable: PermResult; // 禁用和启用用户
  func_admin_user_password_reset: PermResult; // 重置用户密码
  func_admin_user_edit: PermResult; // 设置角色等

  func_channel_create_public: PermResult; // 频道 - 创建公开频道
  func_channel_create_non_public: PermResult; // 频道 - 创建个人频道
  // gChannelSettingsEdit: PermResult; // 频道管理 - 通用权限
}

export type PermCheckKey = keyof SystemRolePermSheet
