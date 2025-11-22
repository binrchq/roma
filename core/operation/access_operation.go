package operation

import (
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"gorm.io/gorm"
)

type AccessOperation struct {
	DB *gorm.DB
}

func NewAccessOperation() *AccessOperation {
	return &AccessOperation{DB: global.GetDB()}
}

func NewAccessOperationWithDebug() *AccessOperation {
	return &AccessOperation{DB: global.GetDB().Debug()}
}

func NewAccessOperationWithDB(db *gorm.DB) *AccessOperation {
	return &AccessOperation{DB: db}
}

func (a *AccessOperation) GetAccessLogs(username string, resourceType string, limit int) ([]*model.AccessLog, error) {
	logs := []*model.AccessLog{}
	query := a.DB.Order("timestamp DESC")

	// 如果提供了用户名，需要通过 user_id 关联查询
	if username != "" {
		var user model.User
		if err := a.DB.Where("username = ?", username).First(&user).Error; err == nil {
			query = query.Where("user_id = ?", user.ID)
		} else {
			// 如果用户不存在，返回空结果
			return []*model.AccessLog{}, nil
		}
	}

	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func (a *AccessOperation) GetCredentialLogs(username string, limit int) ([]*model.CredentialAccessLog, error) {
	logs := []*model.CredentialAccessLog{}
	query := a.DB.Order("timestamp DESC")

	// 如果提供了用户名，需要通过 user_id 关联查询
	if username != "" {
		var user model.User
		if err := a.DB.Where("username = ?", username).First(&user).Error; err == nil {
			query = query.Where("user_id = ?", user.ID)
		} else {
			// 如果用户不存在，返回空结果
			return []*model.CredentialAccessLog{}, nil
		}
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
