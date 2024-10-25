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
	db.Unscoped().Delete(&model.UserEmojiModel{}, "id = ?", c.Query("id"))
	return nil
}

func UserEmojiList(c *fiber.Ctx) error {
	ui := getCurUser(c)

	return utils.APIPaginatedList(c, func(page, pageSize int) ([]*model.UserEmojiModel, int64, error) {
		return model.UserEmojiList(ui.ID, 1, -1)
	})
}
