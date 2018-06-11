package data

import (
	"time"
)

type User struct {
	ID            uint64 `gorm:"primary_key"`
	Name          string
	Discriminator string
	Email         string `gorm:"type:varchar(254);unique_index"`
	Token         string
	RefreshToken  string
	Expiry        time.Time
	//OwnedBots          []Bot    `gorm:"many2many:user_bots;"`
	//Reviews            []Review `gorm:"manymany:user_reviews;"`
	//Invited            []Bot    `gorm:"many2one:user_reviews;"`
	SessionToken       string // TODO wat?
	SessionTokenSecret string // TODO wat?
	Agreed             bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time `sql:"index"`
}
