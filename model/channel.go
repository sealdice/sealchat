package model

import (
	"fmt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
	"sealchat/protocol"
	"strings"
	"time"
)

type ChannelModel struct {
	StringPKBaseModel
	Name         string `json:"name"`
	ParentID     string `json:"parentId" gorm:"null"` // 好像satori协议这里不统一啊
	MembersCount int    `json:"membersCount" gorm:"-"`
	IsPrivate    bool   `json:"isPrivate" gorm:"index"` // 是私聊频道吗？
	RecentSentAt int64  `json:"recentSentAt"`           // 最近发送消息的时间
}

func (*ChannelModel) TableName() string {
	return "channels"
}

func (m *ChannelModel) UpdateRecentSent() {
	m.RecentSentAt = time.Now().UnixMilli()
	db.Model(m).Update("recent_sent_at", m.RecentSentAt)
}

func (c *ChannelModel) GetPrivateUserIDs() []string {
	return strings.SplitN(c.ID, ":", 2)
}

func (c *ChannelModel) ToProtocolType() *protocol.Channel {
	channelType := protocol.TextChannelType
	if c.IsPrivate {
		channelType = protocol.DirectChannelType
	}
	return &protocol.Channel{
		ID:   c.ID,
		Name: c.Name,
		Type: channelType,
	}
}

const (
	ChannelPermUserALL = "@all"
)

type ChannelPermModel struct {
	StringPKBaseModel
	ChannelID string `json:"channel_id" gorm:"index"` // 准入的频道ID
	UserID    string `json:"user_id" gorm:"index"`    // 准入的用户ID
}

func (*ChannelPermModel) TableName() string {
	return "channel_perms"
}

func ChannelPublicNew(channelID string, name string) *ChannelModel {
	ch := &ChannelModel{StringPKBaseModel: StringPKBaseModel{ID: channelID}, Name: name}
	db.Create(ch)
	db.Create(&ChannelPermModel{StringPKBaseModel: StringPKBaseModel{ID: gonanoid.Must()}, ChannelID: channelID, UserID: ChannelPermUserALL})
	return ch
}

func ChannelPrivateNew(userID1, userID2 string) (ch *ChannelModel, isNew bool) {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}

	chId := fmt.Sprintf("%s:%s", userID1, userID2)

	chExists := &ChannelModel{}
	db.Where("id = ?", chId).First(&chExists)
	if chExists.ID != "" {
		return chExists, false
	}

	ch = &ChannelModel{StringPKBaseModel: StringPKBaseModel{ID: chId}, IsPrivate: true, Name: "@私聊频道"}

	db.Create(ch)
	db.Create(&ChannelPermModel{StringPKBaseModel: StringPKBaseModel{ID: gonanoid.Must()}, ChannelID: chId, UserID: userID1})
	db.Create(&ChannelPermModel{StringPKBaseModel: StringPKBaseModel{ID: gonanoid.Must()}, ChannelID: chId, UserID: userID2})

	return ch, true
}

func ChannelList(userId string) []*ChannelModel {
	// 加载有权限访问的频道
	var items []*ChannelPermModel
	db.Where("user_id = ? or user_id = ?", ChannelPermUserALL, userId).Find(&items)

	ids := lo.Map(items, func(item *ChannelPermModel, index int) string {
		return item.ChannelID
	})
	ids = lo.Uniq(ids)

	var items2 []*ChannelModel
	db.Where("id in ?", ids).Order("is_private asc, created_at asc").Find(&items2)

	// 加载私人频道
	uid2channel := make(map[string]*ChannelModel)

	var uids []string
	for _, i := range items2 {
		if i.IsPrivate {
			for _, j := range i.GetPrivateUserIDs() {
				if j != userId {
					uids = append(uids, j)
					uid2channel[j] = i
				}
			}
			i.Name = "@私聊频道"
		}
	}
	uids = lo.Uniq(uids)

	var uItems []*UserModel
	db.Where("id in ?", uids).Find(&uItems)
	for _, i := range uItems {
		uid2channel[i.ID].Name = fmt.Sprintf("%s", i.Nickname)
	}

	return items2
}
