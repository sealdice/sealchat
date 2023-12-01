package api

import (
	_ "embed"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"io/fs"
	"log"
	"net/http"
	"sealchat/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/afero"
)

var appConfig *utils.AppConfig
var appFs afero.Fs

func Init(config *utils.AppConfig, uiStatic fs.FS) {
	appConfig = config
	corsConfig := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, ChannelId",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           3600,
	})

	appFs = afero.NewOsFs()

	app := fiber.New()
	app.Use(corsConfig)
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())

	// Default /test
	app.Use(config.WebUrl, filesystem.New(filesystem.Config{
		Root:       http.FS(uiStatic),
		PathPrefix: "ui/dist",
		MaxAge:     5 * 60,
	}))

	v1 := app.Group("/api/v1")
	v1.Post("/user/signup", UserSignup)
	v1.Post("/user/signin", UserSignin)

	v1Auth := v1.Group("")
	v1Auth.Use(SignCheckMiddleware)
	v1Auth.Post("/user/change_password", UserChangePassword)
	v1Auth.Get("/user/info", UserInfo)
	v1Auth.Put("/user/info", UserInfoUpdate)

	v1Auth.Get("/timeline/list", TimelineList)

	v1Auth.Post("/upload", Upload)
	v1Auth.Get("/attachments/list", AttachmentList)
	v1Auth.Static("/attachments", "./data/upload")

	v1.Get("/config", func(ctx *fiber.Ctx) error {
		return ctx.JSON(appConfig)
	})

	v1AuthAdmin := v1Auth.Group("")
	v1AuthAdmin.Use(UserRoleAdminMiddleware)
	v1AuthAdmin.Get("/bot_token/list", BotTokenList)
	v1AuthAdmin.Post("/bot_token/add", BotTokenAdd)
	v1AuthAdmin.Delete("/bot_token/:id", BotTokenDelete)
	v1AuthAdmin.Get("/admin/user/list", AdminUserList)
	v1AuthAdmin.Put("/admin/user/disable/:id", AdminUserDisable)
	v1AuthAdmin.Put("/admin/user/enable/:id", AdminUserEnable)
	v1AuthAdmin.Put("/admin/user/reset_password/:id", AdminUserResetPassword)

	v1AuthAdmin.Put("/config", func(ctx *fiber.Ctx) error {
		var newConfig utils.AppConfig
		err := ctx.BodyParser(&newConfig)
		if err != nil {
			return err
		}
		appConfig = &newConfig
		utils.WriteConfig(appConfig)
		return nil
	})

	websocketWorks(app)

	// Default :3212
	log.Fatal(app.Listen(config.ServeAt))
}
