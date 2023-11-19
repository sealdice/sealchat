package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type ApiMsgPayload struct {
	Api  string `json:"api"`
	Echo string `json:"echo"`
}

type ChatContext struct {
	Conn            *websocket.Conn
	User            *model.UserModel
	Echo            string
	ConnMap         *utils.SyncMap[string, *websocket.Conn]
	ChannelUsersMap *utils.SyncMap[string, *utils.SyncSet[string]]
}

func Init() {
	config := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           3600,
	})

	app := fiber.New()
	app.Use(config)
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())

	//app.Get("/test$", func(c *fiber.Ctx) error {
	//	return c.Redirect("/test/")
	//})
	app.Static("/test/", "./static")

	connMap := &utils.SyncMap[string, *websocket.Conn]{}
	channelUsersMap := &utils.SyncMap[string, *utils.SyncSet[string]]{}

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	v1 := app.Group("/api/v1")
	v1.Post("/user/signup", UserSignup)
	v1.Post("/user/signin", UserSignin)

	v1Auth := v1.Group("")
	v1Auth.Use(SignCheckMiddleware)
	v1Auth.Post("/user/change_password", UserChangePassword)
	v1Auth.Get("/user/info", UserInfo)

	app.Get("/ws/seal", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		sessionId := c.Cookies("session")

		if sessionId == "" {
			sessionId = gonanoid.Must()
			//c.WriteMessage(websocket.BinaryMessage, []byte(sessionId))
		}

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt      int
			msg     []byte
			err     error
			curUser *model.UserModel
		)

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				// 解析错误
				break
			}

			solved := false
			gatewayMsg := protocol.GatewayPayloadStructure{}
			err := json.Unmarshal(msg, &gatewayMsg)
			if err == nil {
				// 信令
				switch gatewayMsg.Op {
				case protocol.OpIdentify:
					fmt.Println("新客户端接入")
					if gatewayMsg.Body != nil {
						fmt.Println("gatewayMsg.Body", gatewayMsg.Body)
						// 有身份信息
						if m, ok := gatewayMsg.Body.(map[string]interface{}); ok {
							if tokenAny, exists := m["token"]; exists {
								if token, ok := tokenAny.(string); ok {
									user, err := model.UserVerifyAccessToken(token)
									if err == nil {
										connMap.Store(user.ID, c)
										curUser = user
										utils.Must0(c.WriteJSON(protocol.GatewayPayloadStructure{
											Op: protocol.OpReady,
											Body: map[string]interface{}{
												"user": curUser,
											},
										}))
										solved = true
										break
									}
								}
							}
						}
					}

					utils.Must0(c.WriteJSON(protocol.GatewayPayloadStructure{
						Op: protocol.OpReady,
						Body: map[string]interface{}{
							"errorMsg": "no auth",
						},
					}))
					solved = true
				case protocol.OpPing:
					fmt.Println("ping")
					utils.Must0(c.WriteJSON(protocol.GatewayPayloadStructure{
						Op: protocol.OpPong,
					}))
					solved = true
				}
			}

			if !solved {
				apiMsg := ApiMsgPayload{}
				err := json.Unmarshal(msg, &apiMsg)
				ctx := &ChatContext{
					Conn:            c,
					User:            curUser,
					Echo:            apiMsg.Echo,
					ConnMap:         connMap,
					ChannelUsersMap: channelUsersMap,
				}

				if err == nil {
					switch apiMsg.Api {
					case "channel.create":
						apiChannelCreate(c, msg, apiMsg.Echo)
						solved = true
					case "channel.list":
						apiChannelList(ctx, msg)
						solved = true
					case "channel.enter":
						apiChannelEnter(ctx, msg)
						solved = true
					// case "guild.list":
					//	 apiChannelList(c, msg, apiMsg.Echo)
					//	 solved = true
					case "message.create":
						apiMessageCreate(ctx, msg)
						solved = true
					case "message.list":
						apiMessageList(ctx, msg)
						solved = true
					}
				}
			}

			log.Printf("recv: %s  %d", msg, mt)
			//if err = c.WriteMessage(mt, msg); err != nil {
			//	log.Println("write:", err)
			//	break
			//}
		}
	}))

	log.Fatal(app.Listen(":3212"))
}
