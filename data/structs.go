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
	Collections        []*Collection `gorm:"many2many:user_collections"`
	SessionToken       string        // TODO wat?
	SessionTokenSecret string        // TODO wat?
	Agreed             bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time `sql:"index"`
}

type Collection struct {
	ID        uint64 `gorm:"primary_key"`
	Name      string
	Owner     *User    `gorm:"many2many:user_collections;"`
	Tracks    []*Track `gorm:"many2many:collection_tracks;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type Track struct {
	ID         uint64 `gorm:"primary_key"`
	Name       string
	Collection *Collection `gorm:"many2many:collection_tracks;"`
	URL        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `sql:"index"`
}
