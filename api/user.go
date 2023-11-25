package api

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"regexp"
	"sealchat/model"
	"strings"
)

func SignCheckMiddleware(c *fiber.Ctx) error {
	//token := c.Cookies("token")
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

	user, err := model.UserVerifyAccessToken(token)
	if err != nil {
		//fmt.Println(err.Error())
		//return c.Redirect("http://127.0.0.1:4455/login", http.StatusMovedPermanently)
		return &fiber.Error{
			Code:    http.StatusUnauthorized,
			Message: "凭证错误，需要重新登录",
		}
	}
	//if !*resp.Active {
	//	//return c.Redirect("http://127.0.0.1:4455/login", http.StatusMovedPermanently)
	//	fmt.Println("过期了")
	//	return &fiber.Error{
	//		Code:    http.StatusUnauthorized,
	//		Message: "凭证过期，需要重新登录",
	//	}
	//} else {
	c.Locals("user", user)

	if isWriteCookie {
		c.Cookie(&fiber.Cookie{
			Name:   "Authorization",
			Value:  token,
			MaxAge: 3600, // 存活时间为3600秒
		})
	}

	//}
	model.TimelineUpdate(user.ID)

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
	oldPassword := c.FormValue("old_password")
	newPassword := c.FormValue("new_password")
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
	return c.JSON(fiber.Map{
		"message": "修改密码成功",
	})
}

func UserInfo(c *fiber.Ctx) error {
	u := getCurUser(c)
	return c.JSON(fiber.Map{
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

	db := model.GetDB()
	u2 := &model.UserModel{}
	db.Select("id").Where("nickname = ?", data.Nickname).First(&u2)
	if u2.ID != "" {
		c.Status(http.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "昵称已被占用",
		})
	}

	u := getCurUser(c)
	u.Nickname = data.Nickname
	u.Brief = data.Brief
	u.SaveInfo()

	return c.JSON(fiber.Map{
		"user": u,
	})
}
