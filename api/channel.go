package api

import (
	"github.com/gofiber/fiber/v2"

	"sealchat/model"
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
