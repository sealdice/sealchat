package model

import (
	"fmt"
	"strings"
	"time"

	"sealchat/protocol"
)

type ChannelModel struct {
	StringPKBaseModel
	Name         string `json:"name"`
	Note         string `json:"note"`                   // 这是一份注释，用于管理人员辨别数据
	RootId       string `json:"rootId"`                 // 如果未来有多级子频道，那么rootId指向顶层
	ParentID     string `json:"parentId" gorm:"null"`   // 好像satori协议这里不统一啊
	IsPrivate    bool   `json:"isPrivate" gorm:"index"` // 是私聊频道吗？
	RecentSentAt int64  `json:"recentSentAt"`           // 最近发送消息的时间
	UserID       string `json:"userId"`                 // 创建者ID
	PermType     string `json:"permType"`               // public 公开 non-public 非公开 private 私聊

	FriendInfo   *FriendModel `json:"friendInfo,omitempty" gorm:"-"`
	MembersCount int          `json:"membersCount" gorm:"-"`
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

func ChannelPublicNew(channelID string, ch *ChannelModel, creatorId string) *ChannelModel {
	ch.ID = channelID
	ch.UserID = creatorId

	db.Create(ch)
	return ch
}

func ChannelPrivateNew(userID1, userID2 string) (ch *ChannelModel, isNew bool) {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}

	chId := fmt.Sprintf("%s:%s", userID1, userID2)

	u1 := UserGet(userID1)
	u2 := UserGet(userID2)

	if u1 == nil || u2 == nil {
		return nil, false
	}

	chExists := &ChannelModel{}
	db.Where("id = ?", chId).Limit(1).Find(&chExists)
	if chExists.ID != "" {
		return chExists, false
	}

	ch = &ChannelModel{
		StringPKBaseModel: StringPKBaseModel{ID: chId},
		IsPrivate:         true,
		Name:              "@私聊频道",
		PermType:          "private",
		Note:              fmt.Sprintf("%s-%s", u1.Username, u2.Username),
	}
	db.Create(ch)

	return ch, true
}

// ChannelGet 获取频道
func ChannelGet(id string) (*ChannelModel, error) {
	var item ChannelModel
	err := db.Limit(1).Find(&item, "id = ?", id).Error
	return &item, err
}

func ChannelPrivateList(userId string) []*ChannelModel {
	// 加载有权限访问的频道
	var items []*ChannelModel
	q := db.Where("is_private = true and ", true).Order("created_at asc")
	q.Find(&items)

	return items
}
