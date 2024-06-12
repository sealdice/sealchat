package api

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
)

type ApiMsgPayload struct {
	Api  string `json:"api"`
	Echo string `json:"echo"`
}

type WsSyncConn struct {
	*websocket.Conn
	Mux sync.RWMutex
}

func (c *WsSyncConn) WriteJSON(v interface{}) error {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return c.Conn.WriteJSON(v)
}

type ConnInfo struct {
	User         *model.UserModel
	Conn         *WsSyncConn
	LastPingTime int64
	ChannelId    string
}

var commandTips utils.SyncMap[string, map[string]string]

func websocketWorks(app *fiber.App) {
	channelUsersMap := &utils.SyncMap[string, *utils.SyncSet[string]]{}
	userId2ConnInfo := &utils.SyncMap[string, *utils.SyncMap[*WsSyncConn, *ConnInfo]]{}

	clientEnter := func(c *WsSyncConn, body any) (curUser *model.UserModel, curConnInfo *ConnInfo) {
		if body != nil {
			// 有身份信息
			m, ok := body.(map[string]any)
			if !ok {
				return nil, nil
			}
			tokenAny, exists := m["token"]
			if !exists {
				return nil, nil
			}
			token, ok := tokenAny.(string)
			if !ok {
				return nil, nil
			}

			var user *model.UserModel
			var err error

			if len(token) == 32 {
				user, err = model.BotVerifyAccessToken(token)
			} else {
				user, err = model.UserVerifyAccessToken(token)
			}

			if err == nil {
				m, _ := userId2ConnInfo.LoadOrStore(user.ID, &utils.SyncMap[*WsSyncConn, *ConnInfo]{})
				curConnInfo = &ConnInfo{Conn: c, LastPingTime: time.Now().Unix(), User: user}
				m.Store(c, curConnInfo)

				curUser = user
				_ = c.WriteJSON(protocol.GatewayPayloadStructure{
					Op: protocol.OpReady,
					Body: map[string]any{
						"user": curUser,
					},
				})
				return
			}
		}

		_ = c.WriteJSON(protocol.GatewayPayloadStructure{
			Op: protocol.OpReady,
			Body: map[string]any{
				"errorMsg": "no auth",
			},
		})
		return nil, nil
	}

	go func() {
		// 持续删除超时连接
		// for {
		//	time.Sleep(5 * time.Second)
		//	now := time.Now().Unix()
		//	oldLen := connMap.Len()
		//	connMap.Range(func(key string, value *ConnInfo) bool {
		//		if now-value.LastPingTime > 20 {
		//			_ = value.Conn.Close()
		//			connMap.Delete(key)
		//
		//			channelUsersMap.Range(func(chId string, value *utils.SyncSet[string]) bool {
		//				value.Delete(key)
		//				return true
		//			})
		//		}
		//		return true
		//	})
		//
		//	if connMap.Len()-oldLen != 0 {
		//		ctx := &ChatContext{
		//			ConnMap:         connMap,
		//			ChannelUsersMap: channelUsersMap,
		//		}
		//		ctx.BroadcastEvent(&protocol.Event{
		//			Type: "channel-updated",
		//		})
		//	}
		// }
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

	app.Get("/ws/seal", websocket.New(func(rawConn *websocket.Conn) {
		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt          int
			msg         []byte
			err         error
			curUser     *model.UserModel
			curConnInfo *ConnInfo
		)
		c := &WsSyncConn{rawConn, sync.RWMutex{}}

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
					curUser, curConnInfo = clientEnter(c, gatewayMsg.Body)
					if curUser == nil {
						_ = c.Close()
						return
					}
					solved = true
				case protocol.OpPing:
					if curUser == nil {
						solved = true
						continue
					}
					if info, ok := userId2ConnInfo.Load(curUser.ID); ok {
						if info2, ok := info.Load(c); ok {
							info2.LastPingTime = time.Now().Unix()
						}
					}
					_ = c.WriteJSON(protocol.GatewayPayloadStructure{
						Op: protocol.OpPong,
					})
					solved = true
				}
			}

			if !solved {
				apiMsg := ApiMsgPayload{}
				err := json.Unmarshal(msg, &apiMsg)

				var members []*model.MemberModel
				db := model.GetDB()
				db.Where("user_id = ?", curUser.ID).Find(&members)

				ctx := &ChatContext{
					Conn:            c,
					User:            curUser,
					Echo:            apiMsg.Echo,
					ConnInfo:        curConnInfo,
					Members:         members,
					ChannelUsersMap: channelUsersMap,
					UserId2ConnInfo: userId2ConnInfo,
				}

				if err == nil {
					switch apiMsg.Api {
					case "channel.create":
						apiChannelCreate(ctx, msg, apiMsg.Echo)
						solved = true
					case "channel.private.create":
						// 私聊
						apiChannelPrivateCreate(ctx, msg)
						solved = true
					case "channel.list":
						apiChannelList(ctx, msg)
						solved = true
					case "channel.members_count":
						apiChannelMemberCount(ctx, msg)
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
					case "message.delete":
						apiMessageDelete(ctx, msg)
						solved = true
					case "message.list":
						apiMessageList(ctx, msg)
						solved = true
					case "guild.member.list":
						apiGuildMemberList(ctx, msg)
						solved = true
					case "bot.info.set_name":
						apiBotInfoSetName(ctx, msg)
						solved = true
					case "bot.command.register":
						apiBotCommandRegister(ctx, msg)
						solved = true
					case "bot.channel_member.set_name":
						apiBotChannelMemberSetName(ctx, msg)
					}
				}
			}

			log.Printf("recv: %s  %d", msg, mt)
			// if err = c.WriteMessage(mt, msg); err != nil {
			//	log.Println("write:", err)
			//	break
			// }
		}

		// 连接断开，后续封装成函数
		userId2ConnInfo.Range(func(key string, value *utils.SyncMap[*WsSyncConn, *ConnInfo]) bool {
			exists := value.Delete(c)
			if exists {
				return false
			}
			return true
		})
		channelUsersMap.Range(func(chId string, value *utils.SyncSet[string]) bool {
			value.Delete(curUser.ID)
			return true
		})
	}))
}
