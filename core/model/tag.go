package model

// Tag 标签
type Tag struct {
	ID        int64  `gorm:"column:id;primaryKey" json:"id"`      // 标签的唯一标识，作为主键
	Label     string `gorm:"column:label" json:"name"`            // 标签的名称
	Value     string `gorm:"column:value" json:"value"`           // 标签的值
	CreatedAt string `gorm:"column:created_at" json:"created_at"` // 标签创建时间
	UpdatedAt string `gorm:"column:updated_at" json:"updated_at"` // 标签更新时间
}
