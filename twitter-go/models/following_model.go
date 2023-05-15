package models

import (
	"github.com/google/uuid"
	// "gorm.io/gorm"
)


type Following struct{
	// gorm.Model
	ID uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	FollowerId uuid.UUID `json:"follower_id" gorm:"type:uuid"` // who follows
	FolloweeId uuid.UUID `json:"followee_id" gorm:"type:uuid"` // who is being followed
}