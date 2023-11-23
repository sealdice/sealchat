package api

import (
	"github.com/gofiber/contrib/websocket"
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
	"time"
)

type ChatContext struct {
	Conn            *websocket.Conn
	User            *model.UserModel
	Echo            string
	ConnMap         *utils.SyncMap[string, *ConnInfo]
	ChannelUsersMap *utils.SyncMap[string, *utils.SyncSet[string]]
}

func (ctx *ChatContext) BroadcastEvent(data *protocol.Event) {
	data.Timestamp = time.Now().Unix()
	ctx.ConnMap.Range(func(key string, value *ConnInfo) bool {
		utils.Must0(value.Conn.WriteJSON(struct {
			protocol.Event
			Op protocol.Opcode `json:"op"`
		}{
			// 协议规定: 事件中必须含有 channel，message，user
			Event: *data,
			Op:    protocol.OpEvent,
		}))
		return true
	})
}

func (ctx *ChatContext) BroadcastEventInChannel(channelId string, data *protocol.Event) {
	data.Timestamp = time.Now().Unix()
	if lst, exists := ctx.ChannelUsersMap.Load(channelId); exists {
		lst.Range(func(key string) bool {
			if value, exists := ctx.ConnMap.Load(key); exists {
				utils.Must0(value.Conn.WriteJSON(struct {
					protocol.Event
					Op protocol.Opcode `json:"op"`
				}{
					// 协议规定: 事件中必须含有 channel，message，user
					Event: *data,
					Op:    protocol.OpEvent,
				}))
			}
			return true
		})
	}
}
