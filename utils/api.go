package utils

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
)

// APIPaginatedList 处理获取频道角色列表的请求
func APIPaginatedList[T any](c *fiber.Ctx, queryFunc func(page, pageSize int) ([]*T, int64, error)) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "40"))
	if err != nil || pageSize < 1 {
		pageSize = 40
	}

	items, total, err := queryFunc(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "获取列表失败",
			"err":     err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
