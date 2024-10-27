package model

import (
	"sealchat/utils"
)

type TimelineUserLastRecordModel struct {
	StringPKBaseModel
	LastBeMentionedId string `json:"lastBeMentionedId"` // 被提及ID
	LastSysMessageId  string `json:"lastSysMessageId"`
}

func (*TimelineUserLastRecordModel) TableName() string {
	return "timeline_user_last_record"
}

func TimelineUpdate(userId string) {
	item := &TimelineUserLastRecordModel{}
	db.Where("id = ?", userId).Limit(1).Find(&item)

	// 如果没有记录，就创建一条
	if item.ID == "" {
		item.ID = userId
		db.Create(&item)
	}

	changed := false
	changed2 := false
	item.LastBeMentionedId, changed2 = updateTimelineByMention(userId, item.LastBeMentionedId)
	changed = changed || changed2

	if changed {
		db.Save(&item)
	}
}

func updateTimelineByMention(userId, lastBeMentionedId string) (string, bool) {
	var item MentionModel
	db.Where("id = ?", lastBeMentionedId).Limit(1).Find(&item)

	var items []MentionModel
	createdAt := item.CreatedAt
	db.Where("receiver_id = ? and created_at > ?", userId, createdAt).Order("created_at asc").Find(&items)

	newItems := []*TimelineModel{}
	for _, i := range items {
		newItems = append(newItems, &TimelineModel{
			StringPKBaseModel: StringPKBaseModel{
				ID: utils.NewID(),
			},
			Type:        "mention",
			LocPostType: i.LocPostType,
			LocPostID:   i.LocPostID,

			SenderId:   i.SenderId,
			ReceiverId: i.ReceiverId,
		})
	}

	if len(items) > 0 {
		// TODO: 后面sort一下再统一插入
		db.CreateInBatches(newItems, 50)
		return items[len(items)-1].ID, true
	}

	return "", false
}

type TimelineModel struct {
	StringPKBaseModel
	Type   string `gorm:"not null" json:"type"` // 消息类型
	Title  string `json:"title"`
	Brief  string `json:"brief"`
	UserID string `gorm:"index" json:"userId"` // 发起人

	ReceiverId string `gorm:"index" json:"receiverId"` // 被提及人
	SenderId   string `gorm:"index" json:"senderId"`   // 发起人

	LocPostType string `gorm:"index" json:"locPostType"` // 事发地对象类型
	LocPostID   string `gorm:"index" json:"locPostId"`   // 事发地对象ID

	RelatedType string `gorm:"index" json:"relatedType"` // 消息类型
	RelatedID   string `gorm:"index" json:"relatedId"`   // 消息ID

	Data   ByteArray `gorm:"null" json:"data"`    // 附加数据，一般用不着
	IsRead bool      `gorm:"index" json:"isRead"` // 是否已读
}

func (*TimelineModel) TableName() string {
	return "timeline"
}
