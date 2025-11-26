package operation

import (
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"gorm.io/gorm"
)

type ResourceRoleOperation struct {
	DB *gorm.DB
}

func NewResourceRoleOperation() *ResourceRoleOperation {
	return &ResourceRoleOperation{DB: global.GetDB()}
}

func (r *ResourceRoleOperation) AssignRoleToResource(resourceID int64, resourceType string, roleID uint, spaceID *uint) error {
	rr := &model.ResourceRole{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		RoleID:       roleID,
		SpaceID:      spaceID,
	}
	return r.DB.Create(rr).Error
}

func (r *ResourceRoleOperation) GetResourceRoles(resourceID int64, resourceType string) ([]*model.ResourceRole, error) {
	var roles []*model.ResourceRole
	if err := r.DB.Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Preload("Role").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *ResourceRoleOperation) RemoveRoleFromResource(resourceID int64, resourceType string, roleID uint) error {
	return r.DB.Where("resource_id = ? AND resource_type = ? AND role_id = ?", resourceID, resourceType, roleID).
		Delete(&model.ResourceRole{}).Error
}

func (r *ResourceRoleOperation) GetResourcesByRole(roleID uint) ([]*model.ResourceRole, error) {
	var roles []*model.ResourceRole
	if err := r.DB.Where("role_id = ?", roleID).
		Preload("Role").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
