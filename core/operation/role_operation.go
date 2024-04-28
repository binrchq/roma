package operation

import (
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"gorm.io/gorm"
)

type RoleOperation struct {
	DB *gorm.DB
}

func NewRoleOperation() *RoleOperation {
	return &RoleOperation{DB: global.GetDB()}
}

func NewRoleOperationWithDebug() *RoleOperation {
	return &RoleOperation{DB: global.GetDB().Debug()}
}
func NewRoleOperationWithDB(db *gorm.DB) *RoleOperation {
	return &RoleOperation{DB: db}
}

func (r *RoleOperation) GetAllRoles() ([]model.Role, error) {
	var roles []model.Role
	// 获取所有角色
	if err := r.DB.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleOperation) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	// 根据名称获取角色
	if err := r.DB.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleOperation) Create(role *model.Role) (*model.Role, error) {
	// 创建角色
	if err := r.DB.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleOperation) Update(role *model.Role) (*model.Role, error) {
	// 更新角色
	if err := r.DB.Save(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}
