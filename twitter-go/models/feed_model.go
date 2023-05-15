package models

import (
	"time"

	"github.com/google/uuid"
	// "gorm.io/gorm"
)

type Feed struct {
	// gorm.Model
	ID uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	UserId uuid.UUID `json:"userId" gorm:"type:uuid"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}


func (f *Feed) TableName() string{
	return "feeds"
}