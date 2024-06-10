package model

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"sealchat/utils"
)

var db *gorm.DB

type StringPKBaseModel struct {
	ID        string     `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt,omitempty"`
	UpdatedAt time.Time  `json:"updatedAt,omitempty"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt,omitempty"`
}

func (m *StringPKBaseModel) Init() {
	id := utils.NewID()
	m.ID = id
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	m.DeletedAt = nil
}

func DBInit() {
	var err error
	db, err = gorm.Open(sqlite.Open("./data/chat.db"), &gorm.Config{})
	// db.Exec("PRAGMA foreign_keys = ON") // 外键约束，不需要
	db.Exec("PRAGMA journal_mode=WAL")

	if err != nil {
		panic("连接数据库失败")
	}
	db.AutoMigrate(&ChannelModel{})
	db.AutoMigrate(&GuildModel{})
	db.AutoMigrate(&MessageModel{})
	db.AutoMigrate(&UserModel{})
	db.AutoMigrate(&AccessTokenModel{})
	db.AutoMigrate(&MemberModel{})
	db.AutoMigrate(&Attachment{})
	db.AutoMigrate(&MentionModel{})
	db.AutoMigrate(&TimelineModel{})
	db.AutoMigrate(&TimelineUserLastRecordModel{})
	db.AutoMigrate(&UserEmojiModel{})
	db.AutoMigrate(&BotTokenModel{})

	isPermTableExists := db.Migrator().HasTable(&ChannelPermModel{})
	db.AutoMigrate(&ChannelPermModel{})

	if !isPermTableExists {
		var items []*ChannelModel
		db.Find(&items)

		for _, i := range items {
			db.Create(&ChannelPermModel{
				StringPKBaseModel: StringPKBaseModel{
					ID: utils.NewID(),
				},
				ChannelID: i.ID,
				UserID:    ChannelPermUserALL,
			})
		}
	}

	// 初始化默认频道
	var channelCount int64
	db.Model(&ChannelModel{}).Count(&channelCount)
	if channelCount == 0 {
		db.Create(&ChannelModel{
			StringPKBaseModel: StringPKBaseModel{
				ID: utils.NewID(),
			},
			Name: "默认",
		})
	}
}

func GetDB() *gorm.DB {
	return db
}

func FlushWAL() {
	_ = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	_ = db.Exec("PRAGMA shrink_memory")
}
