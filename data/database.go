package data

import (
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/SelfBotBot/selfbot/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Handler struct {
	Engine    *gorm.DB
	SQLConfig config.MySQL
}

func NewHandler(sqlConf config.MySQL) (*Handler, error) {

	if !(sqlConf.Enabled) {
		fmt.Println("No SQL conf enabled!")
		return nil, errors.New("no SQL conf enabled")
	}

	var err error

	handler := &Handler{}
	handler.SQLConfig = sqlConf

	handler.Engine, err = gorm.Open("mysql", sqlConf.URI)
	if err != nil {
		return nil, err
	}

	err = handler.Sync()
	if err != nil {
		return nil, err
	}

	return handler, nil

}

func (h *Handler) Sync() error {
	h.Engine.AutoMigrate(&User{}, &Collection{}, &Track{})
	gob.Register(User{})
	gob.Register(Track{})
	gob.Register(Collection{})

	h.Engine.Model(&User{}).Related(&Collection{}, "Collections")

	h.Engine.Model(&Collection{}).Related(&Track{}, "Tracks")
	h.Engine.Model(&Track{}).Related(&Collection{}, "Collection")

	return nil
}
