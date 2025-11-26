package operation

import (
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"gorm.io/gorm"
)

type SpaceOperation struct {
	DB *gorm.DB
}

func NewSpaceOperation() *SpaceOperation {
	return &SpaceOperation{DB: global.GetDB()}
}

// CreateSpace 创建空间（需要 admin 权限）
func (s *SpaceOperation) CreateSpace(space *model.Space) (*model.Space, error) {
	if err := s.DB.Create(space).Error; err != nil {
		return nil, err
	}
	return space, nil
}

// GetSpaceByID 获取空间
func (s *SpaceOperation) GetSpaceByID(id uint) (*model.Space, error) {
	var space model.Space
	if err := s.DB.Preload("Members.User").
		Preload("Members.Role").
		Preload("Resources").
		Preload("Creator").
		First(&space, id).Error; err != nil {
		return nil, err
	}
	return &space, nil
}

// GetSpaceByName 根据名称获取空间
func (s *SpaceOperation) GetSpaceByName(name string) (*model.Space, error) {
	var space model.Space
	if err := s.DB.Where("name = ?", name).First(&space).Error; err != nil {
		return nil, err
	}
	return &space, nil
}

// AddSpaceMember 添加空间成员
func (s *SpaceOperation) AddSpaceMember(spaceID, userID, roleID uint) (*model.SpaceMember, error) {
	member := &model.SpaceMember{
		SpaceID:  spaceID,
		UserID:   userID,
		RoleID:   roleID,
		IsActive: true,
	}
	if err := s.DB.Create(member).Error; err != nil {
		return nil, err
	}
	return member, nil
}

// IsUserInSpace 检查用户是否在空间中
func (s *SpaceOperation) IsUserInSpace(userID, spaceID uint) (bool, error) {
	var count int64
	if err := s.DB.Model(&model.SpaceMember{}).
		Where("user_id = ? AND space_id = ? AND is_active = ?", userID, spaceID, true).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetSpaceMember 获取空间成员
func (s *SpaceOperation) GetSpaceMember(userID, spaceID uint) (*model.SpaceMember, error) {
	var member model.SpaceMember
	if err := s.DB.Where("user_id = ? AND space_id = ? AND is_active = ?", userID, spaceID, true).
		Preload("Role").First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// AssignResourceToSpace 将资源分配到空间
func (s *SpaceOperation) AssignResourceToSpace(spaceID uint, resourceID int64, resourceType string) error {
	rs := &model.ResourceSpace{
		SpaceID:      spaceID,
		ResourceID:   resourceID,
		ResourceType: resourceType,
	}
	return s.DB.Create(rs).Error
}

// GetResourceSpace 获取资源所属的空间
func (s *SpaceOperation) GetResourceSpace(resourceID int64, resourceType string) (*model.ResourceSpace, error) {
	var rs model.ResourceSpace
	if err := s.DB.Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Preload("Space").First(&rs).Error; err != nil {
		return nil, err
	}
	return &rs, nil
}

// RemoveSpaceMember 移除空间成员
func (s *SpaceOperation) RemoveSpaceMember(spaceID, userID uint) error {
	return s.DB.Where("space_id = ? AND user_id = ?", spaceID, userID).
		Delete(&model.SpaceMember{}).Error
}

// GetUserSpaces 获取用户所属的所有空间
func (s *SpaceOperation) GetUserSpaces(userID uint) ([]*model.Space, error) {
	var spaces []*model.Space
	if err := s.DB.Table("spaces").
		Joins("JOIN space_members ON spaces.id = space_members.space_id").
		Where("space_members.user_id = ? AND space_members.is_active = ? AND spaces.is_active = ?", userID, true, true).
		Preload("Members.User").
		Preload("Members.Role").
		Preload("Creator").
		Find(&spaces).Error; err != nil {
		return nil, err
	}
	return spaces, nil
}

// GetAllSpaces 获取所有空间（管理员使用）
func (s *SpaceOperation) GetAllSpaces() ([]*model.Space, error) {
	var spaces []*model.Space
	if err := s.DB.Where("is_active = ?", true).
		Preload("Members.User").
		Preload("Members.Role").
		Preload("Creator").
		Find(&spaces).Error; err != nil {
		return nil, err
	}
	return spaces, nil
}
