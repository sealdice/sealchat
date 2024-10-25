package api

import (
	"net/http"
	"sealchat/service"

	"github.com/gofiber/fiber/v2"

	"sealchat/model"
	"sealchat/pm"
)

func AdminUserList(c *fiber.Ctx) error {
	if !CanWithSystemRole(c, pm.PermFuncAdminUserEdit) {
		return nil
	}

	db := model.GetDB()
	var total int64
	db.Model(&model.UserModel{}).Count(&total)

	// 获取列表
	var items []*model.UserModel
	// offset := (page - 1) * pageSize
	db.Order("created_at asc").
		// Offset(offset).Limit(pageSize).
		// Preload("User", func(db *gorm.DB) *gorm.DB {
		//	return db.Select("id, username")
		// }).
		Find(&items)

	for _, i := range items {
		i.RoleIds, _ = model.UserRoleMappingListByUserID(i.ID, "", "system")
	}

	// 返回JSON响应
	return c.JSON(fiber.Map{
		// "page":     page,
		// "pageSize": pageSize,
		"total": total,
		"items": items,
	})
}

func AdminUserDisable(c *fiber.Ctx) error {
	userId := c.Query("id")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "用户ID不能为空",
		})
	}

	err := model.UserSetDisable(userId, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "禁用用户失败",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "用户已成功禁用",
	})
}

func AdminUserEnable(c *fiber.Ctx) error {
	userId := c.Query("id")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "用户ID不能为空",
		})
	}

	err := model.UserSetDisable(userId, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "启用用户失败",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "用户已成功启用",
	})
}

func AdminUserResetPassword(c *fiber.Ctx) error {
	uid := c.Query("id")
	if uid == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "参数错误",
		})
	}

	err := model.UserUpdatePassword(uid, "123456")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "重置密码失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "禁用成功",
	})
}

func AdminUserRoleLinkByUserId(c *fiber.Ctx) error {
	type RequestBody struct {
		UserId  string   `json:"userId"`
		RoleIds []string `json:"roleIds"`
	}

	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		// 处理解析错误
		return err
	}

	if body.UserId == "" || len(body.RoleIds) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "用户ID和角色ID不能为空",
		})
	}

	if !CanWithSystemRole(c, pm.PermFuncAdminUserEdit) {
		return nil
	}

	_, err := service.UserRoleLink(body.RoleIds, []string{body.UserId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "添加用户角色失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "用户角色已添加",
	})
}

func AdminUserRoleUnlinkByUserId(c *fiber.Ctx) error {
	type RequestBody struct {
		UserId  string   `json:"userId"`
		RoleIds []string `json:"roleIds"`
	}

	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		// 处理解析错误
		return err
	}

	if body.UserId == "" || len(body.RoleIds) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "用户ID和角色ID不能为空",
		})
	}

	if !CanWithSystemRole(c, pm.PermFuncAdminUserEdit) {
		return nil
	}

	_, err := service.UserRoleUnlink(body.RoleIds, []string{body.UserId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "删除用户角色失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "用户角色已成功删除",
	})
}
