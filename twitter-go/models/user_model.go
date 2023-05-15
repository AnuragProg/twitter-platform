package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	Mobile string `json:"mobile"`
	Username string `json:"username"`
	Password string `json:"password"`
	JoinedOn time.Time `json:"joinedOn"`

	Feeds []*Feed `gorm:"foreignKey:UserId" json:"feeds,omitempty"`
	Followers []*Following `gorm:"foreignKey:FollowerId" json:"feedsomitempty"`
}


