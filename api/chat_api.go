package api

import (
	"encoding/json"
	"fmt"
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

func apiChannelCreate(c *WsSyncConn, msg []byte, echo string) {
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

	_ = c.WriteJSON(struct {
		Channel *protocol.Channel `json:"channel"`
		Echo    string            `json:"echo"`
	}{Channel: &protocol.Channel{ID: m.ID, Name: m.Name}, Echo: echo})
}

func apiChannelPrivateCreate(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	data := struct {
		Data struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	if ctx.User.ID == data.Data.UserId {
		_ = ctx.Conn.WriteJSON(struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Echo string `json:"echo"`
		}{Code: http.StatusBadRequest, Msg: "不能和自己创建私聊频道", Echo: ctx.Echo})
		return
	}

	ch, isNew := model.ChannelPrivateNew(ctx.User.ID, data.Data.UserId)
	fmt.Println("111", ch, isNew)

	_ = c.WriteJSON(struct {
		Channel *protocol.Channel `json:"channel"`
		IsNew   bool              `json:"is_new"`
		Echo    string            `json:"echo"`
	}{Channel: ch.ToProtocolType(), IsNew: isNew, Echo: ctx.Echo})
}

func apiChannelList(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	items := model.ChannelList(ctx.User.ID)

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
			if !item.IsPrivate {
				item.MembersCount = x.Len()
			}
		}
	}

	_ = c.WriteJSON(ret)
}

func apiChannelMemberCount(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	data := struct {
		Data struct {
			ChannelIds []string `json:"channel_ids"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	id2count := map[string]int{}
	for _, chId := range data.Data.ChannelIds {
		if strings.Contains(chId, ":") {
			// 私聊跳过
			continue
		}
		if x, exists := ctx.ChannelUsersMap.Load(chId); exists {
			id2count[chId] = x.Len()
		}
	}

	_ = c.WriteJSON(struct {
		Data map[string]int `json:"data"`
		Echo string         `json:"echo"`
	}{Data: id2count, Echo: ctx.Echo})
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

	// 如果有旧的，移除旧的
	if ctx.ConnInfo.ChannelId != "" {
		if s, ok := ctx.ChannelUsersMap.Load(ctx.ConnInfo.ChannelId); ok {
			s.Delete(ctx.User.ID)
		}
	}

	member, err := model.MemberGetByUserIDAndChannelID(ctx.User.ID, channelId, ctx.User.Nickname)
	if err != nil {
		return
	}
	memberPT := member.ToProtocolType()

	// 然后添加新的
	chUserSet, _ := ctx.ChannelUsersMap.LoadOrStore(channelId, &utils.SyncSet[string]{})
	chUserSet.Add(ctx.User.ID)

	ctx.ConnInfo.ChannelId = channelId

	ctx.BroadcastEventInChannel(channelId, &protocol.Event{
		Type:   "channel-entered",
		User:   ctx.User.ToProtocolType(),
		Member: memberPT,
	})

	_ = ctx.Conn.WriteJSON(struct {
		Member *protocol.GuildMember `json:"member"`
		Echo   string                `json:"echo"`
	}{
		Member: memberPT,
		Echo:   ctx.Echo,
	})
}

func apiMessageDelete(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	data := struct {
		Data struct {
			ChannelID string `json:"channel_id"`
			MessageID string `json:"message_id"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	db := model.GetDB()
	item := model.MessageModel{}
	db.Where("channel_id = ? and id = ?", data.Data.ChannelID, data.Data.MessageID).First(&item)
	if item.ID != "" {
		if item.UserID != ctx.User.ID {
			// 失败了
			_ = c.WriteJSON(struct {
				Echo string `json:"echo"`
			}{Echo: ctx.Echo})
			return
		}

		item.IsRevoked = true
		db.Model(&item).Update("is_revoked", true)

		var channel model.ChannelModel
		db.Where("id = ?", data.Data.ChannelID).First(&channel)
		if channel.ID == "" {
			return
		}
		channelData := channel.ToProtocolType()

		ctx.BroadcastEvent(&protocol.Event{
			// 协议规定: 事件中必须含有 channel，message，user
			Type:    protocol.EventMessageDeleted,
			Message: item.ToProtocolType2(channelData),
			Channel: channelData,
			User:    ctx.User.ToProtocolType(),
		})

		_ = c.WriteJSON(struct {
			Success bool   `json:"success"`
			Echo    string `json:"echo"`
		}{Success: true, Echo: ctx.Echo})
		return
	}

	_ = c.WriteJSON(struct {
		Echo string `json:"echo"`
	}{Echo: ctx.Echo})
}

func apiMessageCreate(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	echo := ctx.Echo
	db := model.GetDB()
	data := struct {
		Data struct {
			ChannelID string `json:"channel_id"`
			QuoteID   string `json:"quote_id"`
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

	var channel model.ChannelModel
	db.Where("id = ?", data.Data.ChannelID).First(&channel)
	if channel.ID == "" {
		return
	}
	channelData := channel.ToProtocolType()

	var quote model.MessageModel
	if data.Data.QuoteID != "" {
		db.Where("id = ?", data.Data.QuoteID).First(&quote)
		if quote.ID == "" {
			return
		}
	}

	m := model.MessageModel{
		StringPKBaseModel: model.StringPKBaseModel{
			ID: gonanoid.Must(),
		},
		UserID:    ctx.User.ID,
		ChannelID: data.Data.ChannelID,
		MemberID:  member.ID,
		QuoteID:   data.Data.QuoteID,
		Content:   content,
	}
	rows := db.Create(&m).RowsAffected

	if rows > 0 {
		ctx.TagCheck(data.Data.ChannelID, m.ID, content)
		member.UpdateRecentSent()

		userData := ctx.User.ToProtocolType()
		// channelData := &protocol.Channel{ID: data.Data.ChannelID}

		messageData := &protocol.Message{
			ID:        m.ID,
			Content:   content,
			Channel:   channelData,
			User:      userData,
			Member:    member.ToProtocolType(),
			CreatedAt: time.Now().UnixMilli(), // 跟js相匹配
			Quote:     quote.ToProtocolType2(channelData),
		}

		_ = c.WriteJSON(struct {
			protocol.Message
			Echo string `json:"echo"`
		}{
			Message: *messageData,
			Echo:    echo,
		})

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
		_ = c.WriteJSON(struct {
			ErrStatus int    `json:"errStatus"`
			Echo      string `json:"echo"`
		}{
			ErrStatus: http.StatusInternalServerError,
			Echo:      echo,
		})
	}
}

func apiMessageList(ctx *ChatContext, msg []byte) {
	c := ctx.Conn
	db := model.GetDB()
	data := struct {
		Data struct {
			ChannelID string `json:"channel_id"`
			Next      string `json:"next"`

			// 以下两个字段用于查询某个时间段内的消息，可选
			Type     string `json:"type"` // 查询类型，不填为默认，若time则用下面两个值
			FromTime int64  `json:"from_time"`
			ToTime   int64  `json:"to_time"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	var items []*model.MessageModel

	sql := db.Where("channel_id = ?", data.Data.ChannelID)

	if data.Data.Type == "time" {
		// 如果有这俩，附加一个条件
		if data.Data.FromTime > 0 {
			sql = sql.Where("created_at >= ?", time.UnixMilli(data.Data.FromTime))
		}
		if data.Data.ToTime > 0 {
			sql = sql.Where("created_at <= ?", time.UnixMilli(data.Data.ToTime))
		}
	}

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
			return db.Select("id, nickname, avatar, is_bot")
		}).
		Preload("Quote", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, content, created_at, user_id, is_revoked")
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

	for _, i := range items {
		if i.IsRevoked {
			i.Content = ""
		}
		if i.Quote != nil {
			if i.Quote.IsRevoked {
				i.Quote.Content = ""
			}
		}
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

	_ = c.WriteJSON(ret)
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

	_ = c.WriteJSON(ret)
}

func apiBotInfoSetName(ctx *ChatContext, msg []byte) {
	data := struct {
		Data struct {
			Name  string `json:"name"`
			Brief string `json:"brief"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	ctx.User.Nickname = data.Data.Name
	ctx.User.Brief = data.Data.Brief
	ctx.User.SaveInfo()
	for _, i := range ctx.Members {
		i.Nickname = data.Data.Name
		i.SaveInfo()
		// 广播事件，名字更新了
		ctx.BroadcastEventInChannel(i.ChannelID, &protocol.Event{
			Type:   "channel-member-updated",
			Member: i.ToProtocolType(),
		})
	}

	ret := struct {
		Echo string `json:"echo"`
	}{
		Echo: ctx.Echo,
	}
	_ = ctx.Conn.WriteJSON(ret)
}

func apiBotCommandRegister(ctx *ChatContext, msg []byte) {
	data := struct {
		Data map[string]string `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	commandTips.Store(ctx.User.ID, data.Data)

	ret := struct {
		Echo string `json:"echo"`
	}{
		Echo: ctx.Echo,
	}
	_ = ctx.Conn.WriteJSON(ret)
}

func apiBotChannelMemberSetName(ctx *ChatContext, msg []byte) {
	data := struct {
		Data struct {
			Name      string `json:"name"`
			ChannelId string `json:"channel_id"`
			UserId    string `json:"user_id"`
			//Brief string `json:"brief"`
		} `json:"data"`
	}{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return
	}

	member, err := model.MemberGetByUserIDAndChannelIDBase(data.Data.UserId, data.Data.ChannelId, data.Data.Name, false)
	if member != nil {
		member.Nickname = data.Data.Name
		member.SaveInfo()

		// 广播事件，名字更新了
		ctx.BroadcastEventInChannel(data.Data.ChannelId, &protocol.Event{
			Type:   "channel-member-updated",
			Member: member.ToProtocolType(),
		})
	}

	ret := struct {
		Echo string `json:"echo"`
	}{
		Echo: ctx.Echo,
	}
	_ = ctx.Conn.WriteJSON(ret)
}
