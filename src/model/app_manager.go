package model

import "gorm.io/gorm"

type AppManager struct {
	ID      int
	Appid   string `gorm:"index"`
	Manager string
}

type AppManagerDB struct {
	db *gorm.DB
}

func NewAppManagerDB(db *gorm.DB) *AppManagerDB {
	err := db.AutoMigrate(&AppManager{})
	if err != nil {
		panic(err)
	}
	return &AppManagerDB{db: db}
}

func (a *AppManagerDB) SetAppManager(appid, manager string) error {
	res := a.db.Where("appid=?", appid).Updates(&AppManager{Manager: manager})
	if res.RowsAffected == 0 {
		return a.db.Create(&AppManager{
			ID:      0,
			Appid:   appid,
			Manager: manager,
		}).Error
	} else {
		return res.Error
	}
}

func (a *AppManagerDB) GetAppManagers() (managers []*AppManager) {
	a.db.Find(&managers)
	return
}
