package operation

import (
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"gorm.io/gorm"
)

type HostKeyOperation struct {
	DB *gorm.DB
}

func NewHostKeyOperation() *HostKeyOperation {
	return &HostKeyOperation{DB: global.GetDB()}
}

func (h *HostKeyOperation) HostKeyIsExist() bool {
	return h.DB.First(&model.HostKey{}).RowsAffected > 0
}

func (h *HostKeyOperation) SaveHostKey(privateKey []byte, publicKey []byte) (*model.HostKey, error) {
	var hostKey = &model.HostKey{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	if err := h.DB.Create(&hostKey).Error; err != nil {
		return nil, err
	}
	return hostKey, nil
}


// 获取最新的一个作为当前主机密钥
func (h *HostKeyOperation) GetLatestHostKey() (*model.HostKey, error) {
	var hostKey model.HostKey
	if err := h.DB.Order("id desc").First(&hostKey).Error; err != nil {
		return nil, err
	}
	return &hostKey, nil
}
