package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"sealchat/model"
)

func UserEmojiAdd(c *fiber.Ctx) error {
	db := model.GetDB()
	ui := getCurUser(c)

	var body struct {
		AttachmentId string `json:"attachmentId"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "请求参数错误",
		})
	}

	item := &model.UserEmojiModel{
		AttachmentID: body.AttachmentId,
		UserID:       ui.ID,
	}
	item.Init()
	db.Create(item)
	return c.JSON(fiber.Map{
		"item": item,
	})
}

func UserEmojiDelete(c *fiber.Ctx) error {
	db := model.GetDB()
	item := &model.UserEmojiModel{}
	item.Init()
	db.Create(item)
	return nil
}

func UserEmojiList(c *fiber.Ctx) error {
	// page := c.QueryInt("page", 1)
	// pageSize := c.QueryInt("pageSize", 20)
	db := model.GetDB()

	var total int64
	db.Model(&model.UserEmojiModel{}).Count(&total)

	// 获取列表
	var items []*model.UserEmojiModel
	// offset := (page - 1) * pageSize
	db.Order("created_at asc").
		// Offset(offset).Limit(pageSize).
		// Preload("User", func(db *gorm.DB) *gorm.DB {
		//	return db.Select("id, username")
		// }).
		Find(&items)

	// 返回JSON响应
	return c.JSON(fiber.Map{
		// "page":     page,
		// "pageSize": pageSize,
		"total": total,
		"items": items,
	})
}
