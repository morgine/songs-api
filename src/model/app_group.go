package model

import "gorm.io/gorm"

type AppGroup struct {
	ID       int
	ParentID int `gorm:"index"`
	Name     string
	Color    string
}

type AppGroupTable struct {
	ID      int
	GroupID int    `gorm:"index"`
	Appid   string `gorm:"index"`
}

type AppGroupDB struct {
	db *gorm.DB
}

func NewAppGroup(db *gorm.DB) (*AppGroupDB, error) {
	err := db.AutoMigrate(
		&AppGroup{},
		&AppGroupTable{},
	)
	if err != nil {
		return nil, err
	}
	return &AppGroupDB{
		db: db,
	}, nil
}

func (db *AppGroupDB) GetAppGroups() (gs []*AppGroup, err error) {
	err = db.db.Find(&gs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return gs, nil
}

func (db *AppGroupDB) CreateAppGroup(parentID int, g *AppGroup) error {
	g.ParentID = parentID
	return db.db.Create(g).Error
}

func (db *AppGroupDB) DeleteAppGroup(id int) error {
	err := db.db.Where("group_id=?", id).Delete(&AppGroupTable{}).Error
	if err != nil {
		return err
	}
	return db.db.Delete(&AppGroup{ID: id}).Error
}

func (db *AppGroupDB) GetGroupApps(groupID int, selects []string) (apps []*App, err error) {
	query := db.db.Model(&AppGroupTable{}).Where("group_id=?", groupID).Select("appid")
	m := db.db.Where("appid in (?)", query)
	if len(selects) > 0 {
		m = m.Select(selects)
	}
	err = m.Find(&apps).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return apps, nil
}

func (db *AppGroupDB) SetGroupApp(groupID int, appid string) error {
	return db.db.Create(&AppGroupTable{GroupID: groupID, Appid: appid}).Error
}

func (db *AppGroupDB) DeleteGroupApp(groupID int, appid string) error {
	return db.db.Where("group_id=? AND appid=?", groupID, appid).Delete(&AppGroupTable{}).Error
}
