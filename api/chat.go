package api

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
	"net/http"
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
	"time"
)

func apiChannelCreate(c *websocket.Conn, msg []byte, echo string) {
	db := model.GetDB()
	data := struct {
		// guild_id 字段无意义，因为不可能由客户端提交
		Data protocol.Channel `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	m := model.ChannelModel{}
	m.ID = gonanoid.Must()
	m.Name = data.Data.Name

	db.Create(&m)

	utils.Must0(c.WriteJSON(struct {
		protocol.Channel
		Echo string `json:"echo"`
	}{Channel: protocol.Channel{ID: m.ID, Name: m.Name}, Echo: echo}))
}

func apiChannelList(ctx *ChatContext, msg []byte) {
	db := model.GetDB()
	c := ctx.Conn
	var items []*model.ChannelModel
	db.Find(&items)

	ret := struct {
		Data []*model.ChannelModel `json:"data"`
		Next string                `json:"next"`
		Echo string                `json:"echo"`
	}{
		Data: items,
		Echo: ctx.Echo,
	}

	utils.Must0(c.WriteJSON(ret))
}

func apiMessageCreate(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	echo := ctx.Echo
	db := model.GetDB()
	data := struct {
		Data struct {
			ChannelID string `json:"channel_id"`
			Content   string `json:"content"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}
	content := data.Data.Content

	member, err := model.MemberGetByUserIDAndChannelID(ctx.User.ID, data.Data.ChannelID, ctx.User.Nickname)
	if err != nil {
		return
	}

	m := model.MessageModel{
		StringPKBaseModel: model.StringPKBaseModel{
			ID: gonanoid.Must(),
		},
		UserID:    ctx.User.ID,
		ChannelID: data.Data.ChannelID,
		MemberID:  member.ID,
		Content:   content,
	}
	rows := db.Create(&m).RowsAffected

	if rows > 0 {
		userData := ctx.User.ToProtocolType()
		channelData := &protocol.Channel{ID: data.Data.ChannelID}

		messageData := &protocol.Message{
			ID:      m.ID,
			Content: content,
			Channel: channelData,
			User:    userData,
			Member:  member.ToProtocolType(),
		}

		utils.Must0(c.WriteJSON(struct {
			protocol.Message
			Echo string `json:"echo"`
		}{
			Message: *messageData,
			Echo:    echo,
		}))

		// 发出广播事件
		ctx.ConnMap.Range(func(key string, value *websocket.Conn) bool {
			utils.Must0(value.WriteJSON(struct {
				protocol.Event
				Op protocol.Opcode `json:"op"`
			}{
				// 协议规定: 事件中必须含有 channel，message，user
				Event: protocol.Event{
					Type:      protocol.EventMessageCreated,
					Timestamp: time.Now().Unix(),
					Message:   messageData,
					Channel:   channelData,
					User:      userData,
				},
				Op: protocol.OpEvent,
			}))
			return true
		})
	} else {
		utils.Must0(c.WriteJSON(struct {
			ErrStatus int    `json:"errStatus"`
			Echo      string `json:"echo"`
		}{
			ErrStatus: http.StatusInternalServerError,
			Echo:      echo,
		}))
	}
}

func apiMessageList(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	db := model.GetDB()
	data := struct {
		Data struct {
			ChannelID string `json:"channel_id"`
			Next      string `json:"next"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	var items []*model.MessageModel
	// TODO: 分页，next为created_at的时间戳转hex
	db.Where("channel_id = ?", data.Data.ChannelID).
		Order("created_at asc").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Preload("Member", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, channel_id")
		}).Find(&items)

	ret := struct {
		Data []*model.MessageModel `json:"data"`
		Next string                `json:"next"`
		Echo string                `json:"echo"`
	}{
		Data: items,
		Echo: ctx.Echo,
	}

	utils.Must0(c.WriteJSON(ret))
	//utils.Must0(c.WriteJSON(struct {
	//	ErrStatus int    `json:"errStatus"`
	//	Echo      string `json:"echo"`
	//}{
	//	ErrStatus: http.StatusInternalServerError,
	//	Echo:      echo,
	//}))
}
