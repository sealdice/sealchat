package api

import (
	"encoding/hex"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
)

func Upload(c *fiber.Ctx) error {
	// 解析表单中的文件
	form, err := c.MultipartForm()
	if err != nil {
		return err
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

		_ = tempFile.Close()
		err = appFs.Rename(tempFile.Name(), "./assets/upload/"+hexString)
		if err != nil {
			return err
		}

		filenames = append(filenames, hexString)
	}

	return c.JSON(fiber.Map{
		"message": "上传成功",
		"files":   filenames,
	})
}
