package operation

import (
	"time"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"gorm.io/gorm"
)

type BlacklistOperation struct {
	DB *gorm.DB
}

func NewBlacklistOperation() *BlacklistOperation {
	return &BlacklistOperation{DB: global.GetDB()}
}

// CreateOrUpdate 创建或更新黑名单记录
// 输入: blacklist - 黑名单记录
// 输出: *model.Blacklist - 创建或更新后的记录；error - 错误信息
// 必要性: 统一管理黑名单，支持数据库持久化
func (op *BlacklistOperation) CreateOrUpdate(blacklist *model.Blacklist) (*model.Blacklist, error) {
	var existing model.Blacklist
	err := op.DB.Where("ip = ?", blacklist.IP).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		if err := op.DB.Create(blacklist).Error; err != nil {
			return nil, err
		}
		return blacklist, nil
	} else if err != nil {
		return nil, err
	}

	// 更新现有记录
	blacklist.ID = existing.ID
	if err := op.DB.Model(&existing).Updates(blacklist).Error; err != nil {
		return nil, err
	}
	return blacklist, nil
}

// GetByIP 根据IP获取黑名单记录
// 输入: ip - IP地址
// 输出: *model.Blacklist - 黑名单记录；error - 错误信息
// 必要性: 检查IP是否在黑名单中
func (op *BlacklistOperation) GetByIP(ip string) (*model.Blacklist, error) {
	var blacklist model.Blacklist
	err := op.DB.Where("ip = ?", ip).First(&blacklist).Error
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

// IsBlacklisted 检查IP是否在黑名单中（包括是否过期）
// 输入: ip - IP地址
// 输出: bool - 是否在黑名单中；*model.Blacklist - 黑名单记录（如果在）
// 必要性: 快速检查IP是否应该被封禁
func (op *BlacklistOperation) IsBlacklisted(ip string) (bool, *model.Blacklist) {
	blacklist, err := op.GetByIP(ip)
	if err != nil {
		return false, nil
	}

	// 检查是否已过期
	if blacklist.IsExpired() {
		// 自动删除过期记录
		op.DB.Delete(blacklist)
		return false, nil
	}

	return true, blacklist
}

// GetAll 获取所有黑名单记录
// 输入: limit - 限制数量；offset - 偏移量
// 输出: []*model.Blacklist - 黑名单记录列表；error - 错误信息
// 必要性: 前端显示和管理黑名单
func (op *BlacklistOperation) GetAll(limit, offset int) ([]*model.Blacklist, error) {
	var blacklists []*model.Blacklist
	query := op.DB.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&blacklists).Error; err != nil {
		return nil, err
	}

	// 过滤已过期的记录
	validBlacklists := []*model.Blacklist{}
	for _, bl := range blacklists {
		if !bl.IsExpired() {
			validBlacklists = append(validBlacklists, bl)
		} else {
			// 删除过期记录
			op.DB.Delete(bl)
		}
	}

	return validBlacklists, nil
}

// Delete 删除黑名单记录（解禁）
// 输入: ip - IP地址
// 输出: error - 错误信息
// 必要性: 手动解禁IP
func (op *BlacklistOperation) Delete(ip string) error {
	return op.DB.Where("ip = ?", ip).Delete(&model.Blacklist{}).Error
}

// DeleteByID 根据ID删除黑名单记录
// 输入: id - 记录ID
// 输出: error - 错误信息
// 必要性: 通过ID解禁IP
func (op *BlacklistOperation) DeleteByID(id uint) error {
	return op.DB.Delete(&model.Blacklist{}, id).Error
}

// CleanExpired 清理所有过期的黑名单记录
// 输入: 无
// 输出: int64 - 删除的记录数；error - 错误信息
// 必要性: 定期清理过期记录，保持数据库整洁
func (op *BlacklistOperation) CleanExpired() (int64, error) {
	now := time.Now()
	result := op.DB.Where("ban_until IS NOT NULL AND ban_until < ?", now).Delete(&model.Blacklist{})
	return result.RowsAffected, result.Error
}
