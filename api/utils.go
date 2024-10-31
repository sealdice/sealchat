package api

import (
	"io"
	"mime/multipart"
	"net/http"
	"sealchat/pm/gen"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/mikespook/gorbac"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"golang.org/x/crypto/blake2s"

	"sealchat/pm"
)

var copyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func copyZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := io.CopyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}

func SaveMultipartFile(fh *multipart.FileHeader, fOut afero.File, limit int64) (hashOut []byte, err error) {
	var (
		f multipart.File
	)
	f, err = fh.Open()
	if err != nil {
		return
	}

	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	limitReader := io.LimitReader(f, limit)
	hash := lo.Must(blake2s.New256(nil))
	teeReader := io.TeeReader(limitReader, hash)

	_, err = copyZeroAlloc(fOut, teeReader)
	hashOut = hash.Sum(nil)
	return
}

// Can 检查当前用户是否拥有指定项目的指定权限
func Can(c *fiber.Ctx, chId string, relations ...gorbac.Permission) bool {
	ok := pm.Can(getCurUser(c).ID, chId, relations...)
	if !ok {
		_ = c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "无权限访问"})
	}
	return ok
}

// CanWithSystemRole 检查当前用户是否拥有指定权限
func CanWithSystemRole(c *fiber.Ctx, relations ...gorbac.Permission) bool {
	ok := pm.CanWithSystemRole(getCurUser(c).ID, relations...)
	if !ok {
		_ = c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "无权限访问"})
	}
	return ok
}

// CanWithSystemRole2 检查当前用户是否拥有指定权限
func CanWithSystemRole2(c *fiber.Ctx, userId string, relations ...gorbac.Permission) bool {
	ok := pm.CanWithSystemRole(userId, relations...)
	if !ok {
		_ = c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "无权限访问"})
	}
	return ok
}

// CanWithChannelRole 检查当前用户是否拥有指定项目的指定权限
func CanWithChannelRole(c *fiber.Ctx, chId string, relations ...gorbac.Permission) bool {
	ok := pm.CanWithChannelRole(getCurUser(c).ID, chId, relations...)

	if !ok {
		// 额外检查用户的系统级别权限
		var rootPerm []gorbac.Permission
		for _, i := range relations {
			p := i.ID()
			for key, _ := range gen.PermSystemMap {
				if p == key {
					rootPerm = append(rootPerm, gorbac.NewStdPermission(key))
					break
				}
			}
		}

		userId := getCurUser(c).ID
		ok = pm.CanWithSystemRole(userId, rootPerm...)
	}

	if !ok {
		_ = c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "无权限访问"})
	}
	return ok
}
