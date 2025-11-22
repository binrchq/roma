package operation

import (
	"time"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
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

// 获取所有API Keys
func (a *ApikeyOperation) GetAllApiKeys() ([]*model.Apikey, error) {
	apikeys := []*model.Apikey{}
	if err := a.DB.Find(&apikeys).Error; err != nil {
		return nil, err
	}
	return apikeys, nil
}

// 创建API Key
func (a *ApikeyOperation) Create(apikey *model.Apikey) (*model.Apikey, error) {
	if err := a.DB.Create(apikey).Error; err != nil {
		return nil, err
	}
	return apikey, nil
}

// 检查API Key是否存在
func (a *ApikeyOperation) ApiKeyExists(apikey string) (bool, error) {
	var count int64
	if err := a.DB.Model(&model.Apikey{}).Where("api_key = ?", apikey).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// 检查API Key是否有效
func (a *ApikeyOperation) ApiKeyIsValid(apikey string) (bool, error) {
	apikeyModel := &model.Apikey{}
	if err := a.DB.Where("api_key = ? AND expires_at > ?", apikey, time.Now()).First(apikeyModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// 根据ID获取API Key
func (a *ApikeyOperation) GetApiKeyById(id uint) (*model.Apikey, error) {
	apikey := &model.Apikey{}
	if err := a.DB.First(apikey, id).Error; err != nil {
		return nil, err
	}
	return apikey, nil
}

// 根据ID设置API Key过期
func (a *ApikeyOperation) ExpiresApikeyById(id uint) error {
	expiredTime := time.Now()
	return a.DB.Model(&model.Apikey{}).Where("id = ?", id).Update("expires_at", expiredTime).Error
}

// 根据Key获取API Key
func (a *ApikeyOperation) GetApiKeyByKey(apiKey string) (*model.Apikey, error) {
	apikey := &model.Apikey{}
	if err := a.DB.Where("api_key = ?", apiKey).First(apikey).Error; err != nil {
		return nil, err
	}
	return apikey, nil
}
