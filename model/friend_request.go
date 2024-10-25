package model

import (
	"errors"
	"gorm.io/gorm"
)

// FriendRequestModel 好友邀请表
type FriendRequestModel struct {
	StringPKBaseModel
	SenderID   string `json:"senderId" gorm:"index"`   // 发送者
	ReceiverID string `json:"receiverId" gorm:"index"` // 接收者
	Note       string `json:"note"`                    // 申请理由
	Status     string `json:"status" gorm:"index"`     // 可能的值：pending, accept, reject

	UserInfoSender   *UserModel `json:"userInfoSender" gorm:"-"`
	UserInfoReceiver *UserModel `json:"userInfoReceiver" gorm:"-"`
}

func (*FriendRequestModel) TableName() string {
	return "friend_requests"
}

// FriendRequestCreate 创建好友邀请
func FriendRequestCreate(invite *FriendRequestModel) error {
	if invite.Status == "" {
		invite.Status = "pending"
	}

	fr, _ := FriendRequestGetBySenderReceiverStatus(invite.SenderID, invite.ReceiverID, invite.Status)
	if fr.ID != "" {
		return errors.New("exists")
	}

	if fr := FriendRelationGet(invite.SenderID, invite.ReceiverID); fr.IsFriend {
		return errors.New("already friend")
	}

	return db.Create(invite).Error
}

// FriendRequestGetBySenderReceiverStatus 根据发送者、接收者和状态查询一条好友请求记录
func FriendRequestGetBySenderReceiverStatus(senderID, receiverID, status string) (*FriendRequestModel, error) {
	var friendRequest FriendRequestModel
	result := db.Where("sender_id = ? AND receiver_id = ? AND status = ?", senderID, receiverID, status).
		Limit(1).Find(&friendRequest)
	if result.Error != nil {
		return &friendRequest, result.Error
	}
	if result.RowsAffected == 0 {
		return &friendRequest, gorm.ErrRecordNotFound
	}
	return &friendRequest, nil
}

func FriendRequestSetApprove(ID string, approve bool) bool {
	newStatus := "accept"

	if !approve {
		newStatus = "reject"
	}

	q := db.Model(&FriendRequestModel{}).Where("id = ? and status = ?", ID, "pending").
		Updates(map[string]any{"status": newStatus})

	return q.RowsAffected > 0
}

// FriendRequestGetByID 通过ID获取好友邀请
func FriendRequestGetByID(id string) (*FriendRequestModel, error) {
	var item FriendRequestModel
	err := db.Where("id = ?", id).Limit(1).Find(&item).Error
	return &item, err
}

// FriendRequestDelete 删除好友邀请
func FriendRequestDelete(id string) error {
	return db.Where("id = ?", id).Delete(&FriendRequestModel{}).Error
}

// FriendRequestListBySenderID 列出发送者的所有好友邀请
func FriendRequestListBySenderID(senderID string) ([]*FriendRequestModel, error) {
	var invites []*FriendRequestModel
	err := db.Where("sender_id = ? and status = ?", senderID, "pending").Find(&invites).Error
	return invites, err
}

// FriendRequestListByReceiverID 列出接收者的所有好友邀请
func FriendRequestListByReceiverID(receiverID string) ([]*FriendRequestModel, error) {
	var invites []*FriendRequestModel
	err := db.Where("receiver_id = ? and status = ?", receiverID, "pending").Find(&invites).Error
	return invites, err
}
