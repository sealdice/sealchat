package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"net/http"
	"sealchat/model"
	"sealchat/protocol"
	"sealchat/utils"
	"strconv"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	ds "github.com/sealdice/dicescript"
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
		ctx.TagCheck(data.Data.ChannelID, m.ID, content)

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

		if len(content) >= 2 && (content[0] == '/' || content[0] == '.') && content[1] == 'x' {
			vm := ds.NewVM()
			var botText string
			expr := strings.TrimSpace(content[2:])

			if expr == "" {
				expr = "d100"
			}

			err := vm.Run(expr)
			vm.Config.EnableDiceWoD = true
			vm.Config.EnableDiceCoC = true
			vm.Config.EnableDiceFate = true
			vm.Config.EnableDiceDoubleCross = true
			vm.Config.DefaultDiceSideExpr = "面数 ?? 100"
			vm.Config.OpCountLimit = 30000

			if err != nil {
				botText = "出错:" + err.Error()
			} else {
				sb := strings.Builder{}
				sb.WriteString(fmt.Sprintf("算式: %s\n", expr))
				sb.WriteString(fmt.Sprintf("过程: %s\n", vm.Detail))
				sb.WriteString(fmt.Sprintf("结果: %s\n", vm.Ret.ToString()))
				sb.WriteString(fmt.Sprintf("栈顶: %d 层数:%d 算力: %d\n", vm.StackTop(), vm.Depth(), vm.NumOpCount))
				sb.WriteString(fmt.Sprintf("注: 这是一只小海豹，只有基本骰点功能，完整功能请接入海豹核心"))
				botText = sb.String()
			}

			m := model.MessageModel{
				StringPKBaseModel: model.StringPKBaseModel{
					ID: gonanoid.Must(),
				},
				UserID:    "BOT:1000",
				ChannelID: data.Data.ChannelID,
				MemberID:  "BOT:1000",
				Content:   botText,
			}
			db.Create(&m)

			userData := &protocol.User{
				ID:     "BOT:1000",
				Nick:   "小海豹",
				Avatar: "",
				IsBot:  true,
			}
			messageData := m.ToProtocolType2(channelData)
			messageData.User = userData
			messageData.Member = &protocol.GuildMember{
				Name: userData.Nick,
				Nick: userData.Nick,
			}

			ctx.BroadcastEvent(&protocol.Event{
				// 协议规定: 事件中必须含有 channel，message，user
				Type:    protocol.EventMessageCreated,
				Message: messageData,
				Channel: channelData,
				User:    userData,
			})
		}
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
			return db.Select("id, nickname, avatar")
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

func apiGuildMemberList(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	db := model.GetDB()
	data := struct {
		Data struct {
			GuildId string `json:"guild_id"`
			Next    string `json:"next"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	var items []*model.UserModel

	sql := db.Select("nickname, is_bot, avatar, id")

	var count int64
	if data.Data.Next != "" {
		t, err := strconv.ParseInt(data.Data.Next, 36, 64)
		if err != nil {
			return
		}

		sql = sql.Where("created_at < ?", time.UnixMilli(t))
	}

	sql.Order("created_at desc").Find(&items)

	sql.Count(&count)
	var next string

	items = lo.Reverse(items)
	if count > int64(len(items)) && len(items) > 0 {
		next = strconv.FormatInt(items[0].CreatedAt.UnixMilli(), 36)
	}

	ret := struct {
		Data []*model.UserModel `json:"data"`
		Next string             `json:"next"`
		Echo string             `json:"echo"`
	}{
		Data: items,
		Next: next,
		Echo: ctx.Echo,
	}

	utils.Must0(c.WriteJSON(ret))
}
