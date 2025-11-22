package operation

import (
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"gorm.io/gorm"
)

type AuditOperation struct {
	DB *gorm.DB
}

func NewAuditOperation() *AuditOperation {
	return &AuditOperation{DB: global.GetDB()}
}

func NewAuditOperationWithDB(db *gorm.DB) *AuditOperation {
	return &AuditOperation{DB: db}
}

// CreateAuditLog 创建审计日志
func (a *AuditOperation) CreateAuditLog(auditLog *model.AuditLog) error {
	if err := a.DB.Create(auditLog).Error; err != nil {
		return err
	}
	return nil
}

// GetAuditLogs 获取审计日志列表
func (a *AuditOperation) GetAuditLogs(page, pageSize int, filters map[string]interface{}) ([]*model.AuditLog, int64, error) {
	var auditLogs []*model.AuditLog
	var total int64

	query := a.DB.Model(&model.AuditLog{})

	// 应用过滤条件
	if username, ok := filters["username"]; ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username.(string)+"%")
	}
	if action, ok := filters["action"]; ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if actionType, ok := filters["action_type"]; ok && actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if resourceType, ok := filters["resource_type"]; ok && resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&auditLogs).Error; err != nil {
		return nil, 0, err
	}

	return auditLogs, total, nil
}

// GetAuditLogByID 根据ID获取审计日志
func (a *AuditOperation) GetAuditLogByID(id uint) (*model.AuditLog, error) {
	auditLog := &model.AuditLog{}
	if err := a.DB.First(auditLog, id).Error; err != nil {
		return nil, err
	}
	return auditLog, nil
}

