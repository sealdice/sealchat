package api

import (
	"github.com/gofiber/fiber/v2"
	"sealchat/model"
)

func TimelineList(c *fiber.Ctx) error {
	//page := c.QueryInt("page", 1)
	//pageSize := c.QueryInt("pageSize", 20)
	db := model.GetDB()

	var total int64
	db.Model(&model.TimelineModel{}).Count(&total)

	// 获取列表
	var items []model.TimelineModel
	//offset := (page - 1) * pageSize
	db.Order("created_at desc").
		//Offset(offset).Limit(pageSize).
		//Preload("User", func(db *gorm.DB) *gorm.DB {
		//	return db.Select("id, username")
		//}).
		Find(&items)

	// 返回JSON响应
	return c.JSON(fiber.Map{
		//"page":     page,
		//"pageSize": pageSize,
		"total": total,
		"items": items,
	})

	return nil
}
