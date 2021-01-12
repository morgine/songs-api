package model

import "gorm.io/gorm"

type AppPayOut struct {
	ID     int
	Appid  string `gorm:"index"`
	PayOut float64
}

type AppPayOutDB struct {
	db *gorm.DB
}

func NewAppPayOutDB(db *gorm.DB) *AppPayOutDB {
	err := db.AutoMigrate(&AppPayOut{})
	if err != nil {
		panic(err)
	}
	return &AppPayOutDB{db: db}
}

func (a *AppPayOutDB) SetAppPayOut(appid string, num float64) error {
	res := a.db.Where("appid=?", appid).Updates(&AppPayOut{PayOut: num})
	if res.RowsAffected == 0 {
		return a.db.Create(&AppPayOut{
			ID:     0,
			Appid:  appid,
			PayOut: num,
		}).Error
	} else {
		return res.Error
	}
}

func (a *AppPayOutDB) GetAppPayOut(appids []string) (payouts []*AppPayOut) {
	a.db.Where("appid in (?)", appids).Find(&payouts)
	return
}
