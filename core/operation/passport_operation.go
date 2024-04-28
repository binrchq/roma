package operation

import (
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
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
func (op *PassportOperation) CreatePublicPassport(passport *model.Passport) (*model.Passport, error) {
	if err := op.DB.Create(passport).Error; err != nil {
		return nil, err
	}
	return passport, nil
}

func (op *PassportOperation) GetPassportByType(resourceType string) ([]*model.Passport, error) {
	var passports []*model.Passport
	if err := op.DB.Model(&model.Passport{}).Where("resource_type = ?", resourceType).Find(&passports).Error; err != nil {
		return nil, err
	}
	return passports, nil
}
