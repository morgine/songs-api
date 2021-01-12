package model

import (
	"fmt"
	"gorm.io/gorm"
)

var defaultSubscribeMessageCacheKey = "default"

// 数据库中的数据紧用于配置，对于每个公众号直接一次性应用到 redis 中

type CacheStorage interface {
	// 注意：设置表示有缓存，但有缓存不一定代表一定有值，有可能用户就是想取消默认回复，但缓存不能不要，不然每次都查数据库
	//SetDefaultMessages(msgs *SubscribeMessages) error
	//GetDefaultMessages() (msgs *SubscribeMessages, err error)
	//DelDefaultMessages() error

	// 设置 app 消息缓存
	SetAppMessages(appid string, msgs *SubscribeMessages) error
	// 获得 app 消息缓存
	GetAppMessages(appid string) (msgs *SubscribeMessages, err error)
	DelAppMessages(appid string) error
}

type SubscribeMessageDB struct {
	db         *gorm.DB
	cacheStore CacheStorage
}

func NewSubscribeMessageDB(db *gorm.DB, cacheStorage CacheStorage) *SubscribeMessageDB {
	err := db.AutoMigrate(&SubscribeMessageGroup{}, &SubscribeArticleMessage{}, &SubscribeMiniProgramMessage{}, &AppliedApp{})
	if err != nil {
		panic(err)
	}
	return &SubscribeMessageDB{db: db, cacheStore: cacheStorage}
}

type SubscribeMessageGroup struct {
	ID int
	//IsDefault bool
	Name string
}

//func (db *SubscribeMessageDB) isDefaultGroup(groupID int) (ok bool, err error) {
//	g, err := db.getGroup(groupID)
//	if err != nil {
//		return false, err
//	}
//	return g.IsDefault, nil
//}

func (db *SubscribeMessageDB) getGroup(groupID int) (group *SubscribeMessageGroup, err error) {
	group = &SubscribeMessageGroup{}
	err = db.db.Where("id=?", groupID).First(group).Error
	if err != nil {
		return nil, err
	} else {
		return group, nil
	}
}

func (db *SubscribeMessageDB) GetGroups() (groups []*SubscribeMessageGroup) {
	db.db.Find(&groups)
	return
}

func (db *SubscribeMessageDB) SaveGroup(g *SubscribeMessageGroup) error {
	//if g.IsDefault {
	//	db.db.Model(&SubscribeMessageGroup{}).UpdateColumn("is_default", false)
	//}
	err := db.db.Save(g).Error
	if err != nil {
		return err
	}
	//if g.IsDefault {
	//	return db.cacheStore.DelDefaultMessages()
	//}
	return nil
}

func (db *SubscribeMessageDB) DeleteGroup(id int) error {
	//err := db.delIfIsDefaultMessages(id)
	//if err != nil {
	//	return err
	//}
	if id > 1 {
		return db.db.Delete(&SubscribeMessageGroup{ID: id}).Error
	} else {
		return fmt.Errorf("默认分组 %d 不可删除", id)
	}
}

type SubscribeArticleMessage struct {
	ID          int
	GroupID     int `gorm:"index"`
	Title       string
	Description string
	Url         string
	PicFile     string
}

func (db *SubscribeMessageDB) GetArticleMessages(groupID int) (msgs []*SubscribeArticleMessage, err error) {
	err = db.db.Where("group_id=?", groupID).Find(&msgs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return msgs, nil
}

func (db *SubscribeMessageDB) SaveArticleMessage(msg *SubscribeArticleMessage) error {
	err := db.db.Save(msg).Error
	if err != nil {
		return err
	}
	return db.tryDelDefaultCache(msg.GroupID)
	//return db.delIfIsDefaultMessages(msg.GroupID)
}

//func (db *SubscribeMessageDB) delIfIsDefaultMessages(groupID int) error {
//	isDefaultGroup, err := db.isDefaultGroup(groupID)
//	if err != nil {
//		return err
//	}
//	if isDefaultGroup {
//		return db.cacheStore.DelDefaultMessages()
//	}
//	return nil
//}

func (db *SubscribeMessageDB) DeleteArticleMessage(id int) error {
	article := &SubscribeArticleMessage{}
	err := db.db.Where("id=?", id).First(article).Error
	if err != nil {
		return err
	}
	err = db.tryDelDefaultCache(article.GroupID)
	if err != nil {
		return err
	}
	return db.db.Delete(&SubscribeArticleMessage{ID: id}).Error
}

type SubscribeMiniProgramMessage struct {
	ID                 int
	GroupID            int `gorm:"index"`
	Title              string
	Appid              string
	PagePath           string
	ThumbMediaFilename string
}

func (db *SubscribeMessageDB) GetMiniProgramMessages(groupID int) (msgs []*SubscribeMiniProgramMessage, err error) {
	err = db.db.Where("group_id=?", groupID).Find(&msgs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else {
		return msgs, nil
	}
}

func (db *SubscribeMessageDB) SaveMiniProgramMessage(msg *SubscribeMiniProgramMessage) error {
	err := db.db.Save(msg).Error
	if err != nil {
		return err
	}
	return db.tryDelDefaultCache(msg.GroupID)
	//return db.delIfIsDefaultMessages(msg.GroupID)
}

func (db *SubscribeMessageDB) tryDelDefaultCache(groupID int) error {
	if groupID == 1 {
		return db.cacheStore.DelAppMessages(defaultSubscribeMessageCacheKey)
	}
	return nil
}

func (db *SubscribeMessageDB) DeleteMiniProgramMessage(id int) error {
	msg := &SubscribeMiniProgramMessage{}
	err := db.db.Where("id=?", id).First(msg).Error
	if err != nil {
		return err
	}
	err = db.tryDelDefaultCache(msg.GroupID)
	if err != nil {
		return err
	}
	return db.db.Delete(&SubscribeMiniProgramMessage{ID: id}).Error
}

type SubscribeMessages struct {
	Articles []*SubscribeArticleMessage
	Cards    []*SubscribeMiniProgramMessage
}

func (db *SubscribeMessageDB) getGroupMessages(groupID int) (msgs *SubscribeMessages, err error) {
	msgs = &SubscribeMessages{}
	msgs.Articles, err = db.GetArticleMessages(groupID)
	if err != nil {
		return nil, err
	}
	msgs.Cards, err = db.GetMiniProgramMessages(groupID)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (db *SubscribeMessageDB) GetDefaultSubscribeMessages() (msgs *SubscribeMessages, err error) {
	msgs, err = db.cacheStore.GetAppMessages(defaultSubscribeMessageCacheKey)
	if err != nil {
		return nil, err
	}
	if msgs != nil {
		return msgs, nil
	}
	g := &SubscribeMessageGroup{}
	err = db.db.Where("id=?", 1).First(g).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if g.ID > 0 {
		msgs, err = db.getGroupMessages(g.ID)
	} else {
		msgs = &SubscribeMessages{}
	}
	// 就算没有设置默认分组，也要将数据缓存起来，避免重复查询数据库
	err = db.cacheStore.SetAppMessages(defaultSubscribeMessageCacheKey, msgs)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

type AppliedApp struct {
	ID        int
	GroupName string // 最后一次被应用的分组名称
	Appid     string `gorm:"index"`
}

func (db *SubscribeMessageDB) delAppliedApp(appid string) error {
	return db.db.Where("appid=?", appid).Delete(&AppliedApp{}).Error
}

func (db *SubscribeMessageDB) saveAppliedApp(groupName, appid string) error {
	return db.db.Where("appid=?", appid).FirstOrCreate(&AppliedApp{Appid: appid, GroupName: groupName}).Error
}

func (db *SubscribeMessageDB) ApplyApp(groupID int, appids []string) error {
	if groupID <= 1 {
		return fmt.Errorf("默认分组不可被套用到特定公众号上")
	}
	group, err := db.getGroup(groupID)
	if err != nil {
		return err
	}
	msgs, err := db.getGroupMessages(groupID)
	if err != nil {
		return err
	}
	for _, appid := range appids {
		err = db.cacheStore.SetAppMessages(appid, msgs)
		if err != nil {
			return err
		}
		err = db.saveAppliedApp(group.Name, appid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *SubscribeMessageDB) CancelApp(appids []string) error {
	for _, appid := range appids {
		err := db.cacheStore.DelAppMessages(appid)
		if err != nil {
			return err
		}
		err = db.delAppliedApp(appid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *SubscribeMessageDB) GetAppliedApps() (apps []*AppliedApp, err error) {
	err = db.db.Find(&apps).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return apps, nil
}

func (db *SubscribeMessageDB) GetAppSubscribeMessages(appid string) (msgs *SubscribeMessages, err error) {
	return db.cacheStore.GetAppMessages(appid)
}
