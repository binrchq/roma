package operation

import (
	"errors"
	"fmt"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type UserOperation struct {
	DB *gorm.DB
}

func NewUserOperation() *UserOperation {
	return &UserOperation{DB: global.GetDB()}
}

func NewUserOperationWithDebug() *UserOperation {
	return &UserOperation{DB: global.GetDB().Debug()}
}

func NewUserOperationWithDB(db *gorm.DB) *UserOperation {
	return &UserOperation{DB: db}
}

func (u *UserOperation) AddRoleToUser(userID uint, roleID uint) error {
	var user model.User
	if err := u.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}
	var role model.Role
	if err := u.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	if err := u.DB.Model(&user).Association("Roles").Append(&role); err != nil {
		log.Error().Err(err).Msgf("AddRoleToUser error")
		return errors.New("添加角色失败")
	}

	return nil
}

func (u *UserOperation) CreateUser(user *model.User) (*model.User, error) {
	if err := u.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserOperation) GetUserByID(id uint) (*model.User, error) {
	user := &model.User{}
	if err := u.DB.Preload("Roles").First(user, id).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserOperation) GetUserRoles(userID uint) ([]*model.Role, error) {
	roles := []*model.Role{}
	user := &model.User{}
	if err := u.DB.Preload("Roles").First(user, userID).Error; err != nil {
		return nil, err
	}
	// 将 []Role 转换为 []*model.Role
	for i := range user.Roles {
		roles = append(roles, &user.Roles[i])
	}
	return roles, nil
}

func (u *UserOperation) GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	log.Info().Msgf("username:%s", username)
	if err := u.DB.Preload("Roles").Where("username = ?", username).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserOperation) GetUserRolesByUsername(username string) ([]*model.Role, error) {
	roles := []*model.Role{}
	user := &model.User{}
	if err := u.DB.Where("username = ?", username).First(user).Error; err != nil {
		return nil, err
	}
	//获取关联的角色
	if err := u.DB.Model(user).Association("Roles").Find(&roles); err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateUser 更新用户信息
// 如果密码为空，则不更新密码字段
func (u *UserOperation) UpdateUser(user *model.User) (*model.User, error) {
	// 获取现有用户信息
	existingUser := &model.User{}
	if err := u.DB.First(existingUser, user.ID).Error; err != nil {
		return nil, err
	}

	// 如果密码为空，则不更新密码字段
	if user.Password == "" {
		// 使用 Select 指定要更新的字段，排除密码字段
		if err := u.DB.Model(user).Select("username", "name", "nickname", "email", "public_key", "updated_at").Updates(user).Error; err != nil {
			return nil, err
		}
	} else {
		// 如果提供了密码，需要加密后再更新
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}
		user.Password = hashedPassword
		// 更新包括密码在内的所有字段
		if err := u.DB.Model(user).Select("username", "name", "nickname", "email", "public_key", "password", "updated_at").Updates(user).Error; err != nil {
			return nil, err
		}
	}

	// 重新加载用户信息以返回最新数据
	if err := u.DB.Preload("Roles").First(user, user.ID).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserOperation) DeleteUser(id uint64) error {
	if err := u.DB.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

// 用户禁用
func (u *UserOperation) DisabledUser(id uint64) error {
	if err := u.DB.Model(&model.User{}).Where("id = ?", id).Update("status", 0).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserOperation) GetAllUsers() ([]*model.User, error) {
	users := []*model.User{}
	if err := u.DB.Preload("Roles").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
