package api

import (
	"encoding/json"
	"fmt"
	"log"
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/afero"
)

type ApiMsgPayload struct {
	Api  string `json:"api"`
	Echo string `json:"echo"`
}

type ConnInfo struct {
	Conn         *websocket.Conn
	LastPingTime int64
}

var appFs afero.Fs

func Init() {
	config := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, channel_id",
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

	//app.Get("/test$", func(c *fiber.Ctx) error {
	//	return c.Redirect("/test/")
	//})
	app.Static("/test/", "./static")

	connMap := &utils.SyncMap[string, *ConnInfo]{}
	channelUsersMap := &utils.SyncMap[string, *utils.SyncSet[string]]{}

	go func() {
		// 持续删除超时连接
		for {
			time.Sleep(5 * time.Second)
			now := time.Now().Unix()
			oldLen := connMap.Len()
			connMap.Range(func(key string, value *ConnInfo) bool {
				if now-value.LastPingTime > 20 {
					_ = value.Conn.Close()
					connMap.Delete(key)

					channelUsersMap.Range(func(chId string, value *utils.SyncSet[string]) bool {
						value.Delete(key)
						return true
					})
				}
				return true
			})

			if connMap.Len()-oldLen != 0 {
				ctx := &ChatContext{
					ConnMap:         connMap,
					ChannelUsersMap: channelUsersMap,
				}
				ctx.BroadcastEvent(&protocol.Event{
					Type: "channel-updated",
				})
			}
		}
	}()

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
	v1Auth.Post("/upload", Upload)
	v1Auth.Static("/attachments", "./assets/upload")

	app.Get("/ws/seal", websocket.New(func(c *websocket.Conn) {
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
										connMap.Store(user.ID, &ConnInfo{Conn: c, LastPingTime: time.Now().Unix()})
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
					if info, ok := connMap.Load(curUser.ID); ok {
						info.LastPingTime = time.Now().Unix()
					}

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
					case "message.create", "qqq.x":
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
