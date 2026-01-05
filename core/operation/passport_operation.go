package operation

import (
	"fmt"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"gorm.io/gorm"
)

type PassportOperation struct {
	DB *gorm.DB
}

func NewPassportOperation() *PassportOperation {
	return &PassportOperation{DB: global.GetDB()}
}

func NewPassportOperationWithDebug() *PassportOperation {
	return &PassportOperation{DB: global.GetDB().Debug()}
}
func NewPassportOperationWithDB(db *gorm.DB) *PassportOperation {
	return &PassportOperation{DB: db}
}

func (op *PassportOperation) GetPassports() ([]*model.Passport, error) {
	var passports []*model.Passport
	if err := op.DB.Find(&passports).Error; err != nil {
		return nil, err
	}
	return passports, nil
}

// 创建一个公共凭证
// 注意：passport字段不使用数据库唯一索引（因为SSH私钥内容过长会超过PostgreSQL索引大小限制）
// 改为在应用层检查唯一性
func (op *PassportOperation) CreatePublicPassport(passport *model.Passport) (*model.Passport, error) {
	// 检查是否已存在相同的passport
	var existing model.Passport
	if err := op.DB.Where("passport = ?", passport.Passport).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("passport already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	if err := op.DB.Create(passport).Error; err != nil {
		return nil, err
	}
	return passport, nil
}

// GetPassportByType 根据资源类型获取Passport（兼容旧接口，返回所有匹配的）
func (op *PassportOperation) GetPassportByType(resourceType string) ([]*model.Passport, error) {
	var passports []*model.Passport
	if err := op.DB.Model(&model.Passport{}).Where("resource_type = ?", resourceType).Find(&passports).Error; err != nil {
		return nil, err
	}
	return passports, nil
}

// GetPassportForResource 根据资源类型、空间ID和角色ID获取Passport
// 检索优先级：
// 1. 同时匹配 resource_type、space_id 和 role_id 的Passport
// 2. 匹配 resource_type 和 space_id 的Passport（role_id为空）
// 3. 匹配 resource_type 和 role_id 的Passport（space_id为空）
// 4. 只匹配 resource_type 的通用Passport（space_id和role_id都为空）
// 返回第一个匹配的Passport，如果没有匹配的则返回nil
func (op *PassportOperation) GetPassportForResource(resourceType string, spaceID *uint, roleID *uint) (*model.Passport, error) {
	var passport model.Passport

	// 优先级1: 同时匹配 space_id 和 role_id
	if spaceID != nil && roleID != nil {
		if err := op.DB.Where("resource_type = ? AND space_id = ? AND role_id = ?", resourceType, *spaceID, *roleID).
			First(&passport).Error; err == nil {
			return &passport, nil
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// 优先级2: 匹配 space_id（role_id为空）
	if spaceID != nil {
		if err := op.DB.Where("resource_type = ? AND space_id = ? AND role_id IS NULL", resourceType, *spaceID).
			First(&passport).Error; err == nil {
			return &passport, nil
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// 优先级3: 匹配 role_id（space_id为空）
	if roleID != nil {
		if err := op.DB.Where("resource_type = ? AND space_id IS NULL AND role_id = ?", resourceType, *roleID).
			First(&passport).Error; err == nil {
			return &passport, nil
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// 优先级4: 通用Passport（space_id和role_id都为空）
	if err := op.DB.Where("resource_type = ? AND space_id IS NULL AND role_id IS NULL", resourceType).
		First(&passport).Error; err == nil {
		return &passport, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 没有找到匹配的Passport
	return nil, nil
}
