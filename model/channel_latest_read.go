package model

import (
	"time"

	"gorm.io/gorm/clause"
)

/*
本来我想了一些复杂的方案，并估算了内存和硬盘的使用
但随后，我意识到并不需要考虑那么多。
*/

type ChannelLatestReadModel struct {
	StringPKBaseModel

	ChannelId string `gorm:"index:idx_channel_user,unique" json:"channelId"`    // 目前仅用于频道ID
	UserId    string `gorm:"index:idx_channel_user,unique;index" json:"userId"` // 用户ID

	MessageId   string
	MessageTime int64

	Mark string `json:"mark"` // 特殊标记
}

func (*ChannelLatestReadModel) TableName() string {
	return "channel_latest_read"
}

func ChannelReadListByUserId(userId string) ([]*ChannelLatestReadModel, error) {
	var records []*ChannelLatestReadModel
	err := db.Where("user_id = ?", userId).Find(&records).Error
	return records, err
}

func ChannelUnreadFetch(userId string) (map[string]int64, error) {
	items, err := ChannelReadListByUserId(userId)
	if err != nil {
		return nil, err
	}

	var chIds []string
	var timeLst []time.Time
	for _, i := range items {
		chIds = append(chIds, i.ChannelId)
		timeLst = append(timeLst, time.UnixMilli(i.MessageTime))
	}

	unreadMap, err := MessagesCountByChannelIDsAfterTime(chIds, timeLst, userId)
	if err != nil {
		return nil, err
	}

	return unreadMap, err
}

func ChannelReadSet(channelId, userId string) error {
	var record ChannelLatestReadModel
	err := db.Where("channel_id = ? AND user_id = ?", channelId, userId).Limit(1).Find(&record).Error
	if err != nil {
		return err
	}
	if record.ID == "" {
		// 记录不存在,创建新记录
		record = ChannelLatestReadModel{
			ChannelId:   channelId,
			UserId:      userId,
			MessageTime: time.Now().UnixMilli(),
		}
		return db.Create(&record).Error
	}

	return db.Model(&ChannelLatestReadModel{}).
		Where("channel_id = ? AND user_id = ?", channelId, userId).
		Updates(map[string]any{
			"message_time": time.Now().UnixMilli(),
		}).Error
}

func ChannelReadInit(channelId, userId string) error {
	return db.Clauses(clause.OnConflict{
		DoNothing: true, // 对应 INSERT OR IGNORE
	}).Create(&ChannelLatestReadModel{
		ChannelId:   channelId,
		UserId:      userId,
		MessageTime: 0,
	}).Error
}

func ChannelReadInitInBatches(channelId string, userIds []string) error {
	models := make([]ChannelLatestReadModel, len(userIds))
	for i, userId := range userIds {
		models[i] = ChannelLatestReadModel{
			ChannelId:   channelId,
			UserId:      userId,
			MessageTime: 0,
		}
	}

	return db.Clauses(clause.OnConflict{
		DoNothing: true, // 对应 INSERT OR IGNORE
	}).CreateInBatches(models, 100).Error
}
