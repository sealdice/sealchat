package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"sealchat/model"
)

func SignCheckMiddleware(c *fiber.Ctx) error {
	var token string

	tokens := c.GetReqHeaders()["Authorization"]
	if len(tokens) > 0 {
		token = tokens[0]
	}

	cookieToken := c.Cookies("Authorization")
	isWriteCookie := token != "" && cookieToken != token

	if token == "" {
		token = cookieToken
	}

	var user *model.UserModel
	var err error

	if len(token) == 32 {
		user, err = model.BotVerifyAccessToken(token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(
				fiber.Map{"message": err.Error()},
			)
		}
	} else {
		user, err = model.UserVerifyAccessToken(token)
		if err != nil {
			fmt.Println("xxxx", token, err.Error())
			// return c.Redirect("http://127.0.0.1:4455/login", http.StatusMovedPermanently)
			return c.Status(http.StatusUnauthorized).JSON(
				fiber.Map{"message": "凭证错误，需要重新登录"},
			)
		}
	}

	if user.Role == "role-disabled" {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"message": "帐号被禁用"},
		)
	}
	// if !*resp.Active {
	//	//return c.Redirect("http://127.0.0.1:4455/login", http.StatusMovedPermanently)
	//	fmt.Println("过期了")
	//	return &fiber.Error{
	//		Code:    http.StatusUnauthorized,
	//		Message: "凭证过期，需要重新登录",
	//	}
	// } else {
	c.Locals("user", user)

	if isWriteCookie {
		c.Cookie(&fiber.Cookie{
			Name:   "Authorization",
			Value:  token,
			MaxAge: 3600, // 存活时间为3600秒
		})
	}

	// }
	model.TimelineUpdate(user.ID)

	return c.Next()
}

func UserRoleAdminMiddleware(c *fiber.Ctx) error {
	user := getCurUser(c)
	fmt.Println("????", user.Role)
	if user.Role != "role-admin" {
		return c.Status(http.StatusForbidden).JSON(
			fiber.Map{"message": "没有权限"},
		)
	}
	return c.Next()
}

func getCurUser(c *fiber.Ctx) *model.UserModel {
	return c.Locals("user").(*model.UserModel)
}

// 注册接口
func UserSignup(c *fiber.Ctx) error {
	type RequestBody struct {
		Username string `json:"username" form:"username" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
		Nickname string `json:"nickname" form:"nickname" binding:"required"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "请求参数错误",
		})
	}

	username := requestBody.Username
	password := requestBody.Password

	if username == "" || password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "用户名或密码不能为空",
		})
	}

	if len(username) < 2 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "用户名长度不能小于2位",
		})
	}

	if len(password) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "密码长度不能小于3位",
		})
	}

	requestBody.Nickname = strings.TrimSpace(requestBody.Nickname)
	if requestBody.Nickname == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "昵称不能为空",
		})
	}

	user, err := model.UserCreate(username, password, requestBody.Nickname)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	token, err := model.UserGenerateAccessToken(user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "生成token失败",
		})
	}

	model.TimelineUpdate(user.ID)

	return c.JSON(fiber.Map{
		"message": "注册成功",
		"token":   token,
	})
}

// 登录接口
func UserSignin(c *fiber.Ctx) error {
	type RequestBody struct {
		Username string `json:"username" form:"username" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "请求参数错误",
		})
	}

	username := requestBody.Username
	password := requestBody.Password
	if username == "" || password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "用户名或密码不能为空",
		})
	}

	if len(password) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "密码长度不能小于3位",
		})
	}

	user, err := model.UserAuthenticate(username, password)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	token, err := model.UserGenerateAccessToken(user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "生成token失败",
		})
	}
	return c.JSON(fiber.Map{
		"message": "登录成功",
		"token":   token,
	})
}

// 修改密码接口
func UserChangePassword(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "未提供token",
		})
	}
	user, err := model.UserVerifyAccessToken(tokenString)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var formData struct {
		Password    string `form:"password" json:"password" binding:"required"`
		PasswordNew string `form:"passwordNew" json:"passwordNew" binding:"required"`
	}
	if err := c.BodyParser(&formData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "请求参数错误",
		})
	}

	oldPassword := formData.Password
	newPassword := formData.PasswordNew

	if oldPassword == "" || newPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "旧密码或新密码不能为空",
		})
	}
	if _, err := model.UserAuthenticate(user.Username, oldPassword); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "旧密码错误",
		})
	}
	if err := model.UserUpdatePassword(user.ID, newPassword); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "修改密码失败",
		})
	}

	err = model.AcessTokenDeleteAllByUserID(user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "删除用户凭证失败",
		})
	}

	token, err := model.UserGenerateAccessToken(user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "生成token失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "修改密码成功",
		"token":   token,
	})
}

func UserInfo(c *fiber.Ctx) error {
	u := getCurUser(c)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"user": u,
	})
}

func UserInfoUpdate(c *fiber.Ctx) error {
	type RequestBody struct {
		Nickname string `json:"nick" form:"nick"`
		Brief    string `json:"brief" form:"brief"`
	}

	var data RequestBody
	err := c.BodyParser(&data)
	if err != nil {
		return err
	}
	data.Nickname = strings.TrimSpace(data.Nickname)
	if len(data.Nickname) > 20 {
		return c.JSON(fiber.Map{
			"message": "昵称不能超过20个字符",
		})
	}
	if len(data.Nickname) < 1 {
		return c.JSON(fiber.Map{
			"message": "昵称不能为空",
		})
	}
	if m, _ := regexp.MatchString(`\s`, data.Nickname); m {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "昵称不能包含空格",
		})
	}

	u := getCurUser(c)
	db := model.GetDB()
	u2 := &model.UserModel{}
	db.Select("id").Where("nickname = ? and id != ?", data.Nickname, u.ID).First(&u2)
	if u2.ID != "" {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "昵称已被占用",
		})
	}

	u.Nickname = data.Nickname
	u.Brief = data.Brief
	u.SaveInfo()

	return c.JSON(fiber.Map{
		"user": u,
	})
}

func AdminUserList(c *fiber.Ctx) error {
	// page := c.QueryInt("page", 1)
	// pageSize := c.QueryInt("pageSize", 20)
	db := model.GetDB()

	var total int64
	db.Model(&model.UserModel{}).Count(&total)

	// 获取列表
	var items []model.UserModel
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

func AdminUserRoleSet(c *fiber.Ctx, role string) error {
	uid := c.Params("id")
	if uid == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "参数错误",
		})
	}

	user := &model.UserModel{}
	db := model.GetDB()
	db.First(user, "id = ?", uid)
	if user.ID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "用户不存在",
		})
	}

	curUser := getCurUser(c)
	if user.ID == curUser.ID {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "无法修改当前账号状态",
		})
	}

	user.Role = role
	db.Save(user)

	return c.JSON(fiber.Map{
		"message": "修改成功",
	})
}

func AdminUserDisable(c *fiber.Ctx) error {
	return AdminUserRoleSet(c, "role-disabled")
}

func AdminUserEnable(c *fiber.Ctx) error {
	return AdminUserRoleSet(c, "")
}

func AdminUserResetPassword(c *fiber.Ctx) error {
	uid := c.Params("id")
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
