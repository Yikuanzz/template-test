package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户模型
type User struct {
	BaseModel
	Username     string `gorm:"uniqueIndex;size:50;not null" json:"username" validate:"required,min=3,max=50"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email" validate:"required,email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	Status       int    `gorm:"default:1;not null" json:"status"` // 1-正常, 0-禁用
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// TODO: 在这里添加更多模型
// 示例:
// type Post struct {
//     BaseModel
//     Title   string `gorm:"size:200;not null" json:"title"`
//     Content string `gorm:"type:text" json:"content"`
//     UserID  uint   `gorm:"not null" json:"user_id"`
//     User    User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
// }
