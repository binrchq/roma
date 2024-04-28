package operation

import (
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"gorm.io/gorm"
)

type ApikeyOperation struct {
	DB *gorm.DB
}

func NewApikeyOperation() *ApikeyOperation {
	return &ApikeyOperation{DB: global.GetDB()}
}

func NewApikeyOperationWithDebug() *ApikeyOperation {
	return &ApikeyOperation{DB: global.GetDB().Debug()}
}

func NewApikeyOperationWithDB(db *gorm.DB) *ApikeyOperation {
	return &ApikeyOperation{DB: db}
}

func (a *ApikeyOperation) GetAllApiKeys() ([]*model.Apikey, error) {
	apikeys := []*model.Apikey{}
	if err := a.DB.Find(&apikeys).Error; err != nil {
		return nil, err
	}
	return apikeys, nil
}

func (a *ApikeyOperation) Create(apikey *model.Apikey) (*model.Apikey, error) {
	if err := a.DB.Create(apikey).Error; err != nil {
		return nil, err
	}
	return apikey, nil
}

func (a *ApikeyOperation) ApiKeyExists(apikey string) (bool, error) {
	var count int64
	if err := a.DB.Model(&model.Apikey{}).Where("api_key = ?", apikey).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
