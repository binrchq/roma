package operation

import (
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"gorm.io/gorm"
)

type TagOperation struct {
	DB *gorm.DB
}

func NewTagOperation() *TagOperation {
	return &TagOperation{DB: global.GetDB()}
}

func NewTagOperationWithDebug() *TagOperation {
	return &TagOperation{DB: global.GetDB().Debug()}
}

func NewTagOperationWithDB(db *gorm.DB) *TagOperation {
	return &TagOperation{DB: db}
}

func (t *TagOperation) Create(tag *model.Tag) (*model.Tag, error) {
	if err := t.DB.Create(tag).Error; err != nil {
		return nil, err
	}
	return tag, nil
}

func (t *TagOperation) Update(tag *model.Tag) (*model.Tag, error) {
	if err := t.DB.Save(tag).Error; err != nil {
		return nil, err
	}
	return tag, nil
}

func (t *TagOperation) Delete(tag *model.Tag) error {
	if err := t.DB.Delete(tag).Error; err != nil {
		return err
	}
	return nil
}

func (t *TagOperation) GetTagById(id uint) (*model.Tag, error) {
	var tag model.Tag
	if err := t.DB.Where("id = ?", id).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (t *TagOperation) GetTagByName(name string) (*model.Tag, error) {
	var tag model.Tag
	if err := t.DB.Where("name = ?", name).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (t *TagOperation) GetTagsByIds(ids []uint) ([]*model.Tag, error) {
	var tags []*model.Tag
	if err := t.DB.Where("id IN (?)", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
