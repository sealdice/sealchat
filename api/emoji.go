package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"sealchat/model"
	"sealchat/utils"
)

func UserEmojiAdd(c *fiber.Ctx) error {
	ui := getCurUser(c)

	var body struct {
		AttachmentId string `json:"attachmentId"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "请求参数错误",
		})
	}

	item, _ := model.UserEmojiCreate(ui.ID, body.AttachmentId)
	return c.JSON(fiber.Map{
		"item": item,
	})
}

func UserEmojiDelete(c *fiber.Ctx) error {
	db := model.GetDB()
	var reqBody struct {
		IDs []string `json:"ids"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "无效的请求参数",
		})
	}
	ids := reqBody.IDs
	if len(ids) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID列表不能为空",
		})
	}
	result := db.Unscoped().Delete(&model.UserEmojiModel{}, "id IN ?", ids)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "删除表情失败",
		})
	}
	return c.JSON(fiber.Map{
		"message": "表情删除成功",
		"count":   result.RowsAffected,
	})
}

func UserEmojiList(c *fiber.Ctx) error {
	ui := getCurUser(c)

	return utils.APIPaginatedList(c, func(page, pageSize int) ([]*model.UserEmojiModel, int64, error) {
		return model.UserEmojiList(ui.ID, 1, -1)
	})
}
