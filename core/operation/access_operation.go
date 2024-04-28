package operation

import (
	"bitrec.ai/roma/core/global"
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
