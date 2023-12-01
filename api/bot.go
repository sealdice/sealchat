package api

import (
	"github.com/gofiber/fiber/v2"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
	"sealchat/model"
	"time"
)

func BotTokenList(c *fiber.Ctx) error {
	//page := c.QueryInt("page", 1)
	//pageSize := c.QueryInt("pageSize", 20)
	db := model.GetDB()

	var total int64
	db.Model(&model.BotTokenModel{}).Count(&total)

	// 获取列表
	var items []model.BotTokenModel
	//offset := (page - 1) * pageSize
	db.Order("created_at asc").
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
}

func BotTokenAdd(c *fiber.Ctx) error {
	type RequestBody struct {
		Name string `json:"name"`
	}
	var data RequestBody
	if err := c.BodyParser(&data); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "请求参数错误",
		})
	}

	db := model.GetDB()

	uid := gonanoid.Must()
	// 创建一个永不可能登录的用户
	user := &model.UserModel{
		StringPKBaseModel: model.StringPKBaseModel{
			ID: uid,
		},
		Role:     "",
		Username: gonanoid.Must(),
		Nickname: data.Name,
		Password: "",
		Salt:     "BOT_SALT",
		IsBot:    true,
	}

	if err := db.Create(user).Error; err != nil {
		return err
	}

	item := &model.BotTokenModel{
		StringPKBaseModel: model.StringPKBaseModel{
			ID: uid,
		},
		Name:      data.Name,
		Token:     gonanoid.Must(32),
		ExpiresAt: time.Now().UnixMilli() + 3*365*24*60*60*1e3, // 3 years
	}

	err := db.Create(item).Error
	if err != nil {
		return err
	}

	return c.JSON(item)
}

func BotTokenDelete(c *fiber.Ctx) error {
	db := model.GetDB()
	err := db.Delete(&model.BotTokenModel{}, "id = ?", c.Params("id")).Error
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "删除成功",
	})
}
