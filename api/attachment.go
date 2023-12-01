package api

import (
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"modernc.org/libc/limits"
	"sealchat/model"
)

func Upload(c *fiber.Ctx) error {
	// 解析表单中的文件
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	var channelId string
	channelIds := c.GetReqHeaders()["Channelid"] // header中只能首字大写
	if len(channelIds) > 0 {
		channelId = channelIds[0]
	}

	// 获取上传的文件切片
	files := form.File["file"]
	filenames := []string{}

	tmpDir := "./data/temp/"
	uploadDir := "./data/upload/"

	// 遍历每个文件
	for _, file := range files {
		//f, err := appFs.Open("./assets/" + file.Filename + ".upload")
		//if err != nil {
		//	return err
		//}
		_ = appFs.MkdirAll(tmpDir, 0644)
		_ = appFs.MkdirAll(uploadDir, 0644)

		tempFile, err := afero.TempFile(appFs, tmpDir, "*.upload")
		if err != nil {
			return err
		}

		limit := appConfig.ImageSizeLimit
		if limit == 0 {
			limit = limits.INT_MAX
		}
		hashCode, err := SaveMultipartFile(file, tempFile, limit)
		if err != nil {
			return err
		}
		hexString := hex.EncodeToString(hashCode)
		fn := fmt.Sprintf("%s_%d", hexString, file.Size)

		_ = tempFile.Close()
		err = appFs.Rename(tempFile.Name(), uploadDir+fn)
		if err != nil {
			return err
		}

		model.AttachmentCreate(&model.Attachment{
			Filename:  file.Filename,
			Size:      file.Size,
			Hash:      hashCode,
			ChannelID: channelId,
			UserID:    getCurUser(c).ID,
		})

		filenames = append(filenames, fn)

		// 特殊值处理
		if channelId == "user-avatar" {
			user := getCurUser(c)
			user.Avatar = "id:" + fn
			user.SaveAvatar()
		}
	}

	return c.JSON(fiber.Map{
		"message": "上传成功",
		"files":   filenames,
	})
}

func AttachmentList(c *fiber.Ctx) error {
	var items []*model.Attachment
	user := getCurUser(c)
	model.GetDB().Where("user_id = ?", user.ID).Select("id, created_at, hash").Find(&items)

	return c.JSON(fiber.Map{
		"message": "ok",
		"data":    items,
	})
}
