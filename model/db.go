package model

import (
	"github.com/glebarez/sqlite"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
	"time"
)

var db *gorm.DB

type BaseModel struct {
	ID        uint64     `gorm:"primary_key;autoIncrement" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt"`
}

type StringPKBaseModel struct {
	ID        string     `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt"`
}

type BytePKBaseModel2 struct {
	ID        []byte     `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt"`
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
					ID: gonanoid.Must(),
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
				ID: gonanoid.Must(),
			},
			Name: "默认",
		})
	}
}

func InitTestDB() *gorm.DB {
	return db
}

func GetDB() *gorm.DB {
	return db
}

func FlushWAL() {
	_ = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	_ = db.Exec("PRAGMA shrink_memory")
}
