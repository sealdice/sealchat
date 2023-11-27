package api

import (
	_ "embed"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"io/fs"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/afero"
)

var appFs afero.Fs

func Init(uiStatic fs.FS) {
	config := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, ChannelId",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           3600,
	})

	appFs = afero.NewOsFs()

	app := fiber.New()
	app.Use(config)
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())

	//app.Static("/test/", "./static")
	app.Use("/test", filesystem.New(filesystem.Config{
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
	v1Auth.Static("/attachments", "./assets/upload")

	websocketWorks(app)

	log.Fatal(app.Listen(":3212"))
}
