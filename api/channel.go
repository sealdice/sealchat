package api

import (
	"github.com/gofiber/fiber/v2"

	"sealchat/model"
	"sealchat/pm"
	"sealchat/utils"
)

func ChannelRoles(c *fiber.Ctx) error {
	channelID := c.Query("id")
	if channelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "缺少频道ID",
		})
	}

	return utils.APIPaginatedList(c, func(page, pageSize int) ([]*model.ChannelRoleModel, int64, error) {
		roles, total, err := model.ChannelRoleList(channelID, page, pageSize)
		return roles, total, err
	})
}

func ChannelMembers(c *fiber.Ctx) error {
	channelID := c.Query("id")
	if channelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "缺少频道ID",
		})
	}

	return utils.APIPaginatedList(c, func(page, pageSize int) ([]*model.UserRoleMappingModel, int64, error) {
		items, total, err := model.UserRoleMappingListByChannelID(channelID, page, pageSize)
		utils.QueryOneToManyMap(model.GetDB(), items, func(i *model.UserRoleMappingModel) []string {
			return []string{i.UserID}
		}, func(i *model.UserRoleMappingModel, x []*model.UserModel) {
			i.User = x[0]
		}, "")
		return items, total, err
	})
}

// ChannelInfoEdit 处理频道信息编辑请求
func ChannelInfoEdit(c *fiber.Ctx) error {
	// 获取频道ID
	channelId := c.Query("id")
	if channelId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "频道ID不能为空",
		})
	}

	// TODO: 这里借一下 PermFuncChannelRoleLink 权限，以处理老频道
	if !CanWithChannelRole(c, channelId, pm.PermFuncChannelManageInfo, pm.PermFuncChannelRoleLink) {
		return nil
	}

	// 解析请求体
	var updates model.ChannelModel
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请求参数解析失败",
		})
	}

	// 调用编辑方法
	if err := model.ChannelInfoEdit(channelId, &updates); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "频道信息更新失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "频道信息更新成功",
	})
}

// ChannelInfoGet 处理获取频道信息请求
func ChannelInfoGet(c *fiber.Ctx) error {
	// 获取频道ID
	channelId := c.Query("id")
	if channelId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "频道ID不能为空",
		})
	}

	// 获取频道信息
	channel, err := model.ChannelGet(channelId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取频道信息失败",
		})
	}

	return c.JSON(fiber.Map{
		"item": channel,
	})
}
