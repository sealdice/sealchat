package api

import (
	"fmt"
	"net/http"
	"sealchat/service"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	ds "github.com/sealdice/dicescript"
	"gorm.io/gorm"

	"sealchat/model"
	"sealchat/pm"
	"sealchat/protocol"
	"sealchat/utils"
)

func apiMessageDelete(ctx *ChatContext, data *struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}) (any, error) {
	db := model.GetDB()
	item := model.MessageModel{}
	db.Where("channel_id = ? and id = ?", data.ChannelID, data.MessageID).Limit(1).Find(&item)
	if item.ID != "" {
		if item.UserID != ctx.User.ID {
			return nil, nil // 失败了
		}

		item.IsRevoked = true
		db.Model(&item).Update("is_revoked", true)

		var channel model.ChannelModel
		db.Where("id = ?", data.ChannelID).Limit(1).Find(&channel)
		if channel.ID == "" {
			return nil, nil
		}
		channelData := channel.ToProtocolType()

		ctx.BroadcastEvent(&protocol.Event{
			// 协议规定: 事件中必须含有 channel，message，user
			Type:    protocol.EventMessageDeleted,
			Message: item.ToProtocolType2(channelData),
			Channel: channelData,
			User:    ctx.User.ToProtocolType(),
		})

		return &struct {
			Success bool `json:"success"`
		}{Success: true}, nil
	}

	return nil, nil
}

func apiMessageCreate(ctx *ChatContext, data *struct {
	ChannelID string `json:"channel_id"`
	QuoteID   string `json:"quote_id"`
	Content   string `json:"content"`
}) (any, error) {
	echo := ctx.Echo
	db := model.GetDB()
	channelId := data.ChannelID

	var privateOtherUser string

	// 权限检查
	if len(channelId) < 30 { // 注意，这不是一个好的区分方式
		// 群内
		if !pm.CanWithChannelRole(ctx.User.ID, channelId, pm.PermFuncChannelTextSend, pm.PermFuncChannelTextSendAll) {
			return nil, nil
		}
	} else {
		// 好友/陌生人
		fr, _ := model.FriendRelationGetByID(channelId)
		if fr.ID == "" {
			return nil, nil
		}

		privateOtherUser = fr.UserID1
		if fr.UserID1 == ctx.User.ID {
			privateOtherUser = fr.UserID2
		}
	}

	content := data.Content
	member, err := model.MemberGetByUserIDAndChannelID(ctx.User.ID, data.ChannelID, ctx.User.Nickname)
	if err != nil {
		return nil, err
	}

	channel, _ := model.ChannelGet(channelId)
	if channel.ID == "" {
		return nil, nil
	}
	channelData := channel.ToProtocolType()

	var quote model.MessageModel
	if data.QuoteID != "" {
		db.Where("id = ?", data.QuoteID).Limit(1).Find(&quote)
		if quote.ID == "" {
			return nil, nil
		}
	}

	m := model.MessageModel{
		StringPKBaseModel: model.StringPKBaseModel{
			ID: utils.NewID(),
		},
		UserID:    ctx.User.ID,
		ChannelID: data.ChannelID,
		MemberID:  member.ID,
		QuoteID:   data.QuoteID,
		Content:   content,

		SenderMemberName: member.Nickname,
	}
	rows := db.Create(&m).RowsAffected

	if rows > 0 {
		ctx.TagCheck(data.ChannelID, m.ID, content)
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

		// 发出广播事件
		ev := &protocol.Event{
			// 协议规定: 事件中必须含有 channel，message，user
			Type:    protocol.EventMessageCreated,
			Message: messageData,
			Channel: channelData,
			User:    userData,
		}
		ctx.BroadcastEventInChannel(data.ChannelID, ev)
		ctx.BroadcastEventInChannelForBot(data.ChannelID, ev)
		// ctx.BroadcastEvent(ev)

		if appConfig.BuiltInSealBotEnable {
			builtinSealBotSolve(ctx, data, channelData)
		}

		// 处理维度提醒和消息通知
		if channel.PermType == "private" {
			// 如果是私聊，使得双方可见
			// ids := channel.GetPrivateUserIDs()
			// model.FriendRelationSetVisible(ids[0], ids[1])
			model.FriendRelationSetVisibleById(channel.ID)
			_ = model.ChannelReadInit(data.ChannelID, privateOtherUser)

			// 发送快速更新通知
			ctx.BroadcastToUserJSON(privateOtherUser, map[string]any{
				"op":        0,
				"type":      "message-created-notice",
				"channelId": data.ChannelID,
			})
		} else {
			// 给当前在线人都通知一遍
			var uids []string
			ctx.UserId2ConnInfo.Range(func(key string, value *utils.SyncMap[*WsSyncConn, *ConnInfo]) bool {
				uids = append(uids, key)
				return true
			})

			// 找出当前频道在线的人
			var uidsOnline []string
			if x, exists := ctx.ChannelUsersMap.Load(data.ChannelID); exists {
				x.Range(func(key string) bool {
					uidsOnline = append(uidsOnline, key)
					return true
				})
			}

			_ = model.ChannelReadInitInBatches(data.ChannelID, uids)
			_ = model.ChannelReadSetInBatch([]string{data.ChannelID}, uidsOnline)

			// 发送快速更新通知
			ctx.BroadcastJSON(map[string]any{
				"op":        0,
				"type":      "message-created-notice",
				"channelId": data.ChannelID,
			}, uidsOnline)
		}

		return messageData, nil
	} else {
		return &struct {
			ErrStatus int    `json:"errStatus"`
			Echo      string `json:"echo"`
		}{
			ErrStatus: http.StatusInternalServerError,
			Echo:      echo,
		}, nil
	}
}

func apiMessageList(ctx *ChatContext, data *struct {
	ChannelID string `json:"channel_id"`
	Next      string `json:"next"`

	// 以下两个字段用于查询某个时间段内的消息，可选
	Type     string `json:"type"` // 查询类型，不填为默认，若time则用下面两个值
	FromTime int64  `json:"from_time"`
	ToTime   int64  `json:"to_time"`
}) (any, error) {
	db := model.GetDB()

	// 权限检查
	channelId := data.ChannelID
	if len(channelId) < 30 { // 注意，这不是一个好的区分方式
		// 群内
		if !pm.CanWithChannelRole(ctx.User.ID, channelId, pm.PermFuncChannelRead, pm.PermFuncChannelReadAll) {
			return nil, nil
		}
	} else {
		// 好友/陌生人
		fr, _ := model.FriendRelationGetByID(channelId)
		if fr.ID == "" {
			return nil, nil
		}
	}

	var items []*model.MessageModel
	q := db.Where("channel_id = ?", data.ChannelID)

	if data.Type == "time" {
		// 如果有这俩，附加一个条件
		if data.FromTime > 0 {
			q = q.Where("created_at >= ?", time.UnixMilli(data.FromTime))
		}
		if data.ToTime > 0 {
			q = q.Where("created_at <= ?", time.UnixMilli(data.ToTime))
		}
	}

	var count int64
	if data.Next != "" {
		t, err := strconv.ParseInt(data.Next, 36, 64)
		if err != nil {
			return nil, err
		}

		q = q.Where("created_at < ?", time.UnixMilli(t))
	}

	q.Order("created_at desc").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, nickname, avatar, is_bot")
		}).
		Preload("Member", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, channel_id")
		}).Limit(30).Find(&items)

	utils.QueryOneToManyMap(model.GetDB(), items, func(i *model.MessageModel) []string {
		return []string{i.QuoteID}
	}, func(i *model.MessageModel, x []*model.MessageModel) {
		i.Quote = x[0]
	}, "id, content, created_at, user_id, is_revoked")

	_ = model.ChannelReadSet(data.ChannelID, ctx.User.ID)

	q.Count(&count)
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

	return &struct {
		Data []*model.MessageModel `json:"data"`
		Next string                `json:"next"`
	}{
		Data: items,
		Next: next,
	}, nil
}

func builtinSealBotSolve(ctx *ChatContext, data *struct {
	ChannelID string `json:"channel_id"`
	QuoteID   string `json:"quote_id"`
	Content   string `json:"content"`
}, channelData *protocol.Channel) {
	content := data.Content
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
			sb.WriteString(fmt.Sprintf("过程: %s\n", vm.GetDetailText()))
			sb.WriteString(fmt.Sprintf("结果: %s\n", vm.Ret.ToString()))
			sb.WriteString(fmt.Sprintf("栈顶: %d 层数:%d 算力: %d\n", vm.StackTop(), vm.Depth(), vm.NumOpCount))
			sb.WriteString(fmt.Sprintf("注: 这是一只小海豹，只有基本骰点功能，完整功能请接入海豹核心"))
			botText = sb.String()
		}

		m := model.MessageModel{
			StringPKBaseModel: model.StringPKBaseModel{
				ID: utils.NewID(),
			},
			UserID:    "BOT:1000",
			ChannelID: data.ChannelID,
			MemberID:  "BOT:1000",
			Content:   botText,
		}
		model.GetDB().Create(&m)

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
}

func apiUnreadCount(ctx *ChatContext, data *struct{}) (any, error) {
	chIds, _ := service.ChannelIdList(ctx.User.ID)
	lst, err := model.ChannelUnreadFetch(chIds, ctx.User.ID)
	if err != nil {
		return nil, err
	}
	return lst, err
}
