package model

import "gorm.io/gorm"

type MaterialImage struct {
	ID      int
	Appid   string `gorm:"index"`
	File    string `gorm:"index"`
	MediaID string
}

type MaterialImageModel struct {
	db *gorm.DB
}

func (mim *MaterialImageModel) GetMediaID() {

}
