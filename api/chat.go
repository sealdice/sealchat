package api

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"net/http"
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
	"strconv"
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

	for _, item := range items {
		if x, exists := ctx.ChannelUsersMap.Load(item.ID); exists {
			item.MembersCount = x.Len()
		}
	}

	utils.Must0(c.WriteJSON(ret))
}

// 进入频道
func apiChannelEnter(ctx *ChatContext, msg []byte) {
	data := struct {
		Data struct {
			ChannelId string `json:"channel_id"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	channelId := data.Data.ChannelId

	ctx.ChannelUsersMap.Range(func(key string, value *utils.SyncSet[string]) bool {
		value.Delete(ctx.User.ID)
		return true
	})

	ids, exists := ctx.ChannelUsersMap.Load(channelId)
	if !exists {
		ids = &utils.SyncSet[string]{}
		ctx.ChannelUsersMap.Store(channelId, ids)
	}
	ids.Add(ctx.User.ID)

	ctx.BroadcastEventInChannel(channelId, &protocol.Event{
		Type: "channel-entered",
		User: ctx.User.ToProtocolType(),
	})

	utils.Must0(ctx.Conn.WriteJSON(struct {
		protocol.Message
		Echo string `json:"echo"`
	}{
		Echo: ctx.Echo,
	}))
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
			ID:        m.ID,
			Content:   content,
			Channel:   channelData,
			User:      userData,
			Member:    member.ToProtocolType(),
			CreatedAt: time.Now().UnixMilli(), // 跟js相匹配
		}

		utils.Must0(c.WriteJSON(struct {
			protocol.Message
			Echo string `json:"echo"`
		}{
			Message: *messageData,
			Echo:    echo,
		}))

		// 发出广播事件
		ctx.BroadcastEvent(&protocol.Event{
			// 协议规定: 事件中必须含有 channel，message，user
			Type:    protocol.EventMessageCreated,
			Message: messageData,
			Channel: channelData,
			User:    userData,
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

	sql := db.Where("channel_id = ?", data.Data.ChannelID)

	var count int64
	if data.Data.Next != "" {
		t, err := strconv.ParseInt(data.Data.Next, 36, 64)
		if err != nil {
			return
		}

		sql = sql.Where("created_at < ?", time.UnixMilli(t))
	}

	sql.Order("created_at desc").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Preload("Member", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, channel_id")
		}).Limit(30).Find(&items)

	sql.Count(&count)
	var next string

	items = lo.Reverse(items)
	if count > int64(len(items)) && len(items) > 0 {
		next = strconv.FormatInt(items[0].CreatedAt.UnixMilli(), 36)
	}

	ret := struct {
		Data []*model.MessageModel `json:"data"`
		Next string                `json:"next"`
		Echo string                `json:"echo"`
	}{
		Data: items,
		Next: next,
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
