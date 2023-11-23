package api

import (
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"sealchat/model"
)

func Upload(c *fiber.Ctx) error {
	// 解析表单中的文件
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	var channelId string
	channelIds := c.GetReqHeaders()["channel_id"]
	if len(channelIds) > 0 {
		channelId = channelIds[0]
	}

	// 获取上传的文件切片
	files := form.File["file"]
	filenames := []string{}

	// 遍历每个文件
	for _, file := range files {
		//f, err := appFs.Open("./assets/" + file.Filename + ".upload")
		//if err != nil {
		//	return err
		//}
		_ = appFs.MkdirAll("./assets/temp/", 0644)
		_ = appFs.MkdirAll("./assets/upload/", 0644)

		tempFile, err := afero.TempFile(appFs, "./assets/temp/", "*.upload")
		if err != nil {
			return err
		}
		//appFs.temp
		hashCode, err := SaveMultipartFile(file, tempFile)
		if err != nil {
			return err
		}
		hexString := hex.EncodeToString(hashCode)
		fn := fmt.Sprintf("%s_%d", hexString, file.Size)

		_ = tempFile.Close()
		err = appFs.Rename(tempFile.Name(), "./assets/upload/"+fn)
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
	}

	return c.JSON(fiber.Map{
		"message": "上传成功",
		"files":   filenames,
	})
}
